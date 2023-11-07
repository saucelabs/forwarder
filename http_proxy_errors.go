// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
)

type denyError struct {
	error
}

// ErrorHeader is the header that is set on error responses with the error message.
const ErrorHeader = "X-Forwarder-Error"

var (
	ErrProxyLocalhost = denyError{errors.New("localhost proxying is disabled")}
	ErrProxyDenied    = denyError{errors.New("proxying denied")}
)

func (hp *HTTPProxy) errorResponse(req *http.Request, err error) *http.Response {
	handlers := []errorHandler{
		handleNetError,
		handleTLSRecordHeader,
		handleTLSCertificateError,
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
		code = http.StatusInternalServerError
		msg = "An unexpected error occurred"
		label = "unexpected_error"
	}

	hp.metrics.error(label)

	var body bytes.Buffer
	body.WriteString(msg)
	body.WriteString("\n")
	body.WriteString(err.Error())
	body.WriteString("\n")

	resp := proxyutil.NewResponse(code, &body, req)
	resp.Header.Set(ErrorHeader, err.Error())
	resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp.ContentLength = int64(body.Len())
	return resp
}

type errorHandler func(*http.Request, error) (int, string, string)

func handleNetError(_ *http.Request, err error) (code int, msg, label string) {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			code = http.StatusGatewayTimeout
			msg = "Timed out connecting to remote host"
		} else {
			code = http.StatusBadGateway
			msg = "Failed to connect to remote host"
		}
		label = "net_" + netErr.Op
	}

	return
}

func handleTLSRecordHeader(_ *http.Request, err error) (code int, msg, label string) {
	var headerErr *tls.RecordHeaderError
	if errors.As(err, &headerErr) {
		code = http.StatusBadGateway
		msg = "TLS handshake failed"
		label = "tls_record_header"
	}

	return
}

func handleTLSCertificateError(_ *http.Request, err error) (code int, msg, label string) {
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		code = http.StatusBadGateway
		msg = "TLS handshake failed"
		label = "tls_certificate"
	}

	return
}

func handleDenyError(req *http.Request, err error) (code int, msg, label string) {
	var denyErr denyError
	if errors.As(err, &denyErr) {
		code = http.StatusForbidden
		msg = fmt.Sprintf("proxying is denied to host %q", req.Host)
		label = "denied"
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

func unauthorizedResponse(req *http.Request) *http.Response {
	resp := proxyutil.NewResponse(http.StatusProxyAuthRequired, nil, req)
	resp.Header.Set("Proxy-Authenticate", `Basic realm="Sauce Labs Forwarder"`)
	return resp
}
