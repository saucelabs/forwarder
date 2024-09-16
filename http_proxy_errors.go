// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/saucelabs/forwarder/internal/martian"
	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
)

type denyError struct {
	error
}

// ErrorHeader is the header that is set on error responses with the error message.
const ErrorHeader = "X-Forwarder-Error"

var (
	ErrProxyAuthentication = errors.New("proxy authentication required")

	ErrProxyLocalhost = denyError{errors.New("localhost proxying is disabled")}
	ErrProxyDenied    = denyError{errors.New("proxying denied")}
)

const skipMetricsLabel = "-"

func (hp *HTTPProxy) errorResponse(req *http.Request, err error) *http.Response {
	handlers := []errorHandler{
		handleWindowsNetError,
		handleNetError,
		handleTLSRecordHeader,
		handleTLSCertificateError,
		handleTLSECHRejectionError,
		handleTLSAlertError,
		handleMartianErrorStatus,
		handleAuthenticationError,
		handleDenyError,
		handleStatusText,
	}

	var (
		code       int
		msg, label string
	)
	for _, h := range handlers {
		code, msg, label = h(req, err)
		if code != 0 {
			break
		}
	}
	if code == 0 {
		hp.log.Debugf("error response: unexpected error type: %T", err)
		code = http.StatusInternalServerError
		msg = "encountered an unexpected error"
		label = "unexpected_error"
	}

	if label != skipMetricsLabel {
		hp.metrics.error(label)
	}

	var body bytes.Buffer
	body.WriteString(hp.config.Name)
	body.WriteString(" ")
	body.WriteString(msg)
	body.WriteString("\n")
	body.WriteString(err.Error())
	body.WriteString("\n")

	resp := proxyutil.NewResponse(code, &body, req)
	if code == http.StatusProxyAuthRequired {
		resp.Header.Set("Proxy-Authenticate", fmt.Sprintf("Basic realm=%q", hp.config.Name))
	}
	resp.Header.Set(ErrorHeader, hp.config.Name+" "+err.Error())
	resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp.ContentLength = int64(body.Len())
	return resp
}

type errorHandler func(*http.Request, error) (int, string, string)

func handleWindowsNetError(req *http.Request, err error) (code int, msg, label string) {
	if runtime.GOOS != "windows" {
		return
	}

	var se *os.SyscallError
	if errors.As(err, &se) {
		if se.Syscall == "wsarecv" || se.Syscall == "wsasend" {
			const WSAENETUNREACH = 10051
			if n := errno(se.Err); n == WSAENETUNREACH {
				code = http.StatusBadGateway
				msg = fmt.Sprintf("failed to connect to remote host %q", req.Host)

				label = "net_"
				if se.Syscall == "wsarecv" {
					label += "read"
				} else {
					label += "write"
				}
			}
		}
	}

	return
}

func errno(v error) uintptr {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Uintptr {
		return uintptr(rv.Uint())
	}
	return 0
}

func handleNetError(req *http.Request, err error) (code int, msg, label string) {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			code = http.StatusGatewayTimeout
			msg = fmt.Sprintf("timed out connecting to remote host %q", req.Host)
		} else {
			code = http.StatusBadGateway
			msg = fmt.Sprintf("failed to connect to remote host %q", req.Host)
		}
		label = "net_" + netErr.Op
	}

	return
}

func handleTLSRecordHeader(req *http.Request, err error) (code int, msg, label string) {
	var headerErr tls.RecordHeaderError
	if errors.As(err, &headerErr) {
		code = http.StatusBadGateway
		msg = fmt.Sprintf("tls handshake failed for host %q: ", req.Host)
		if tlsRecordHeaderLooksLikeHTTP(headerErr.RecordHeader) {
			msg += "scheme mismatch (looks like an HTTP request)"
		} else {
			msg += fmt.Sprintf("record header: %x", headerErr.RecordHeader)
		}
		label = "tls_record_header"
	}

	return
}

func handleTLSCertificateError(req *http.Request, err error) (code int, msg, label string) {
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		code = http.StatusBadGateway
		msg = fmt.Sprintf("tls handshake failed for host %q\n", req.Host)
		msg += fmt.Sprintf("unverified certificates:%s\n", describeCertificates(certErr.UnverifiedCertificates))
		label = "tls_certificate"
	}

	return
}

func handleTLSECHRejectionError(req *http.Request, err error) (code int, msg, label string) {
	var echErr *tls.ECHRejectionError
	if errors.As(err, &echErr) {
		code = http.StatusBadGateway
		msg = fmt.Sprintf("tls handshake failed for host %q", req.Host)
		label = "tls_ech_rejection"
	}

	return
}

func handleTLSAlertError(req *http.Request, err error) (code int, msg, label string) {
	var alertErr tls.AlertError
	if errors.As(err, &alertErr) {
		code = http.StatusBadGateway
		msg = fmt.Sprintf("tls alert for host %q", req.Host)
		label = "tls_alert"
	}

	return
}

func handleMartianErrorStatus(req *http.Request, err error) (code int, msg, label string) {
	var martianErr martian.ErrorStatus
	if errors.As(err, &martianErr) {
		code = martianErr.Status
		msg = fmt.Sprintf("proxy error for host %q", req.Host)
		label = "martian_error"
	}

	return
}

func handleAuthenticationError(req *http.Request, err error) (code int, msg, label string) {
	if errors.Is(err, ErrProxyAuthentication) {
		code = http.StatusProxyAuthRequired
		msg = fmt.Sprintf("proxying is denied to host %q", req.Host)
		label = "proxy_authentication"
	}

	return
}

func handleDenyError(req *http.Request, err error) (code int, msg, label string) {
	var denyErr denyError
	if errors.As(err, &denyErr) {
		code = http.StatusForbidden
		msg = fmt.Sprintf("proxying is denied to host %q", req.Host)
		label = skipMetricsLabel
	}

	return
}

// There is a difference between sending HTTP and HTTPS requests in the presence of an upstream proxy.
// For HTTPS client issues a CONNECT request to the proxy and then sends the original request.
// In case the proxy responds with status code 4XX or 5XX to the CONNECT request, the client interprets it as URL error.
func handleStatusText(req *http.Request, err error) (code int, msg, label string) {
	if req.URL.Scheme == "https" && err != nil {
		for i := 400; i < 600; i++ {
			if err.Error() == http.StatusText(i) {
				return i, err.Error(), "https_status_text"
			}
		}
	}

	return
}

func tlsRecordHeaderLooksLikeHTTP(hdr [5]byte) bool {
	return bytes.HasPrefix(hdr[:], []byte("HTTP")) ||
		bytes.HasPrefix(hdr[:], []byte("GET")) ||
		bytes.HasPrefix(hdr[:], []byte("HEAD")) ||
		bytes.HasPrefix(hdr[:], []byte("POST")) ||
		bytes.HasPrefix(hdr[:], []byte("PUT")) ||
		bytes.HasPrefix(hdr[:], []byte("OPTIO"))
}

func describeCertificates(chain []*x509.Certificate) string {
	var sb strings.Builder
	for _, cert := range chain {
		sb.WriteString("\n- ")
		sb.WriteString(cert.Subject.String())
		sb.WriteString(" (issued by ")
		sb.WriteString(cert.Issuer.String())
		sb.WriteString(")")
	}

	return sb.String()
}
