// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"testing"

	"golang.org/x/net/http2"
)

func newReq(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "http://example.test", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func TestHandleContextDeadlineExceededError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  int
		wantLabel string
	}{
		{"direct", context.DeadlineExceeded, http.StatusGatewayTimeout, "request_ctx_deadline"},
		{"wrapped", fmt.Errorf("dialing: %w", context.DeadlineExceeded), http.StatusGatewayTimeout, "request_ctx_deadline"},
		{"unrelated", io.EOF, 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, _, label := handleContextDeadlineExceededError(newReq(t), tc.err)
			if code != tc.wantCode || label != tc.wantLabel {
				t.Fatalf("got (%d, %q), want (%d, %q)", code, label, tc.wantCode, tc.wantLabel)
			}
		})
	}
}

func TestHandleUpstreamEOFError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  int
		wantLabel string
	}{
		{"eof", io.EOF, http.StatusBadGateway, "upstream_eof"},
		{"unexpected_eof", io.ErrUnexpectedEOF, http.StatusBadGateway, "upstream_eof"},
		{"wrapped_eof", fmt.Errorf("reading body: %w", io.EOF), http.StatusBadGateway, "upstream_eof"},
		{"unrelated", context.Canceled, 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, _, label := handleUpstreamEOFError(newReq(t), tc.err)
			if code != tc.wantCode || label != tc.wantLabel {
				t.Fatalf("got (%d, %q), want (%d, %q)", code, label, tc.wantCode, tc.wantLabel)
			}
		})
	}
}

func TestHandleConnectionClosedError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  int
		wantLabel string
	}{
		{"econnreset", syscall.ECONNRESET, http.StatusBadGateway, "conn_reset"},
		{"econnaborted", syscall.ECONNABORTED, http.StatusBadGateway, "conn_aborted"},
		{"net_errclosed", net.ErrClosed, http.StatusBadGateway, "conn_closed"},
		{"wrapped_econnreset", fmt.Errorf("read tcp: %w", syscall.ECONNRESET), http.StatusBadGateway, "conn_reset"},
		{"unrelated", io.EOF, 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, _, label := handleConnectionClosedError(newReq(t), tc.err)
			if code != tc.wantCode || label != tc.wantLabel {
				t.Fatalf("got (%d, %q), want (%d, %q)", code, label, tc.wantCode, tc.wantLabel)
			}
		})
	}
}

func TestHandleHTTP2Error(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  int
		wantLabel string
	}{
		{"stream", http2.StreamError{StreamID: 1, Code: http2.ErrCodeProtocol}, http.StatusBadGateway, "http2_stream"},
		{"connection", http2.ConnectionError(http2.ErrCodeInternal), http.StatusBadGateway, "http2_connection"},
		{"goaway", &http2.GoAwayError{ErrCode: http2.ErrCodeEnhanceYourCalm}, http.StatusBadGateway, "http2_goaway"},
		{"unrelated", io.EOF, 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, _, label := handleHTTP2Error(newReq(t), tc.err)
			if code != tc.wantCode || label != tc.wantLabel {
				t.Fatalf("got (%d, %q), want (%d, %q)", code, label, tc.wantCode, tc.wantLabel)
			}
		})
	}
}
