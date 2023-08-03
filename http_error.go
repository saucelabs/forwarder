// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
	handlers := []errorHandler{
		handleNetError,
		handleDenyError,
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
	if nerr, ok := cause(err).(net.Error); ok {
		if nerr.Timeout() {
			code = http.StatusGatewayTimeout
			msg = "Timed out connecting to remote host"
		} else {
			code = http.StatusBadGateway
			msg = "Failed to connect to remote host"
		}
	}

	return
}

func handleDenyError(req *http.Request, err error) (code int, msg string) {
	if _, ok := cause(err).(denyError); ok { //nolint:errorlint // makes no sense here
		code = http.StatusBadGateway
		msg = fmt.Sprintf("proxying is denied to host %q", req.Host)
	}

	return
}

func unauthorizedResponse(req *http.Request) *http.Response {
	resp := proxyutil.NewResponse(http.StatusProxyAuthRequired, nil, req)
	resp.Header.Set("Proxy-Authenticate", `Basic realm="Sauce Labs Forwarder"`)
	return resp
}

func cause(err error) error {
	cause := err
	for {
		e := errors.Unwrap(cause)
		if e == nil {
			break
		}
		cause = e
	}
	return cause
}
