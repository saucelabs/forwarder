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
	"sync/atomic"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/log"
)

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

func TestMultiListen(t *testing.T) {
	mlc := &mockListenerCallback{
		t: t,
	}

	l := Listener{
		Address:           "localhost:0",
		OptionalAddresses: []string{"localhost:0"},
		Log:               log.NopLogger,
		Callbacks:         mlc,
	}

	err := l.Listen()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 2; i++ {
			_, err := l.Accept()
			if err != nil {
				t.Errorf("l.Accept(): got %v, want no error", err)
				return
			}
		}
		close(done)
	}()

	for _, ll := range l.listeners {
		_, err := net.Dial("tcp", ll.Addr().String())
		if err != nil {
			t.Fatalf("net.Dial(): got %v, want no error", err)
		}
	}

	<-done

	if mlc.accepts.Load() != 2 {
		t.Fatalf("mlc.accepts.Load(): got %d, want 2", mlc.accepts.Load())
	}
}

type mockListenerCallback struct {
	t       *testing.T
	accepts atomic.Int32
	done    chan struct{}
}

func (m *mockListenerCallback) OnAccept(_ net.Conn) {
	m.accepts.Add(1)
}

func (m *mockListenerCallback) OnBindError(_ string, _ error) {
}

func (m *mockListenerCallback) OnTLSHandshakeError(_ *tls.Conn, err error) {
	if !errors.Is(err, context.DeadlineExceeded) {
		m.t.Errorf("tl.OnTLSHandshakeError(): got %v, want %v", err, context.DeadlineExceeded)
	}
	m.done <- struct{}{}
}
