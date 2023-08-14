// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpexpect

import (
	"crypto/tls"
	"net/http/httptrace"
	"net/textproto"
	"testing"
)

func newTestClientTrace(t *testing.T) *httptrace.ClientTrace {
	t.Helper()
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			t.Logf("GetConn for hostPort: %s", hostPort)
		},
		GotConn: func(info httptrace.GotConnInfo) {
			t.Logf("GotConn: %+v", info)
		},
		PutIdleConn: func(err error) {
			t.Logf("PutIdleConn with error: %v", err)
		},
		GotFirstResponseByte: func() {
			t.Logf("GotFirstResponseByte")
		},
		Got100Continue: func() {
			t.Logf("Got100Continue")
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			t.Logf("Got1xxResponse: code=%d, header=%v", code, header)
			return nil
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			t.Logf("DNSStart: %+v", info)
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			t.Logf("DNSDone: %+v", info)
		},
		ConnectStart: func(network, addr string) {
			t.Logf("ConnectStart: network=%s, addr=%s", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			t.Logf("ConnectDone: network=%s, addr=%s, err=%v", network, addr, err)
		},
		TLSHandshakeStart: func() {
			t.Logf("TLSHandshakeStart")
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				t.Logf("TLSHandshakeDone: state=%+v, err=%v", state, err)
			} else {
				t.Logf("TLSHandshakeDone: state=%+v", state)
			}
		},
		WroteHeaderField: func(key string, value []string) {
			t.Logf("WroteHeaderField: key=%s, value=%v", key, value)
		},
		WroteHeaders: func() {
			t.Logf("WroteHeaders")
		},
		Wait100Continue: func() {
			t.Logf("Wait100Continue")
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			t.Logf("WroteRequest: %+v", info)
		},
	}
}
