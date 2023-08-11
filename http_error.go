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

	"github.com/google/martian/v3/proxyutil"
)

type denyError struct {
	error
}

// ErrorHeader is the header that is set on error responses with the error message.
const ErrorHeader = "X-Forwarder-Error"

var ErrProxyLocalhost = denyError{errors.New("localhost proxying is disabled")}

func errorResponse(req *http.Request, err error) *http.Response {
	handlers := []errorHandler{
		handleNetError,
		handleTLSRecordHeader,
		handleTLSCertificateError,
		handleDenyError,
		handleStatusText,
	}

	var (
		code int
		msg  string
	)
	for _, h := range handlers {
		code, msg = h(req, err)
		if code != 0 {
			break
		}
	}
	if code == 0 {
		code = http.StatusInternalServerError
		msg = "An unexpected error occurred"
	}

	resp := proxyutil.NewResponse(code, bytes.NewBufferString(msg+"\n"), req)
	resp.Header.Set(ErrorHeader, err.Error())
	resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp.ContentLength = int64(len(msg) + 1)
	return resp
}

type errorHandler func(*http.Request, error) (int, string)

func handleNetError(_ *http.Request, err error) (code int, msg string) {
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			code = http.StatusGatewayTimeout
			msg = "Timed out connecting to remote host"
		} else {
			code = http.StatusBadGateway
			msg = "Failed to connect to remote host"
		}
	}

	return
}

func handleTLSRecordHeader(_ *http.Request, err error) (code int, msg string) {
	var headerErr *tls.RecordHeaderError
	if errors.As(err, &headerErr) {
		code = http.StatusBadGateway
		msg = "TLS handshake failed"
	}

	return
}

func handleTLSCertificateError(_ *http.Request, err error) (code int, msg string) {
	var certErr *tls.CertificateVerificationError
	if errors.As(err, &certErr) {
		code = http.StatusBadGateway
		msg = "TLS handshake failed"
	}

	return
}

func handleDenyError(req *http.Request, err error) (code int, msg string) {
	var denyErr denyError
	if errors.As(err, &denyErr) {
		code = http.StatusBadGateway
		msg = fmt.Sprintf("proxying is denied to host %q", req.Host)
	}

	return
}

// There is a difference between sending HTTP and HTTPS requests in the presence of an upstream proxy.
// For HTTPS client issues a CONNECT request to the proxy and then sends the original request.
// In case the proxy responds with status code 4XX or 5XX to the CONNECT request, the client interprets it as URL error.
func handleStatusText(req *http.Request, err error) (code int, msg string) {
	if req.URL.Scheme == "https" && err != nil {
		for i := 400; i < 600; i++ {
			if err.Error() == http.StatusText(i) {
				return i, err.Error()
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
