// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/log"
)

func TestListenerListenOnce(t *testing.T) {
	l := Listener{
		Address:   "localhost:0",
		Log:       log.NopLogger,
		Callbacks: &mockListenerCallback{t: t},
	}

	if err := l.Listen(); err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	if err := l.Listen(); err == nil {
		t.Fatal("l.Listen(): got no error, want error")
	}
}

func TestListenerTLSHandshakeTimeout(t *testing.T) {
	tlsCfg := new(tls.Config)
	if err := (&TLSServerConfig{HandshakeTimeout: 100 * time.Millisecond}).ConfigureTLSConfig(tlsCfg); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})

	l := Listener{
		Address:             "localhost:0",
		Log:                 log.NopLogger,
		TLSConfig:           tlsCfg,
		TLSHandshakeTimeout: 100 * time.Millisecond,
		Callbacks: &mockListenerCallback{
			t:    t,
			done: done,
		},
	}

	err := l.Listen()
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		// Accept won't return.
		_, _ = l.Accept()
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	<-done
}

type mockListenerCallback struct {
	t    *testing.T
	done chan struct{}
}

func (m *mockListenerCallback) OnAccept(_ net.Conn) {
}

func (m *mockListenerCallback) OnTLSHandshakeError(_ *tls.Conn, err error) {
	if !errors.Is(err, context.DeadlineExceeded) {
		m.t.Errorf("tl.OnTLSHandshakeError(): got %v, want %v", err, context.DeadlineExceeded)
	}
	m.done <- struct{}{}
}
