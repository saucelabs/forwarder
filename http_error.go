// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import (
	"bytes"
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
	// Get the root cause of the error.
	rootErr := err
	for {
		e := errors.Unwrap(rootErr)
		if e == nil {
			break
		}
		rootErr = e
	}

	var (
		msg  string
		code int
	)
	if nerr, ok := err.(net.Error); ok { //nolint // switch is not an option for type asserts
		if nerr.Timeout() {
			code = http.StatusGatewayTimeout
			msg = "Timed out connecting to remote host"
		} else {
			code = http.StatusBadGateway
			msg = "Failed to connect to remote host"
		}
	} else if _, ok := err.(denyError); ok { //nolint:errorlint // makes no sense here
		code = http.StatusBadGateway
		msg = fmt.Sprintf("Proxying is denied to host %q", req.Host)
	} else {
		code = http.StatusInternalServerError
		msg = "An unexpected error occurred"
	}

	resp := proxyutil.NewResponse(code, bytes.NewBufferString(msg+"\n"), req)
	resp.Header.Set(ErrorHeader, err.Error())
	resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp.ContentLength = int64(len(msg) + 1)
	return resp
}

func unauthorizedResponse(req *http.Request) *http.Response {
	resp := proxyutil.NewResponse(http.StatusProxyAuthRequired, nil, req)
	resp.Header.Set("Proxy-Authenticate", `Basic realm="Sauce Labs Forwarder"`)
	return resp
}
