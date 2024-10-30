// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package conntrack

import (
	"crypto/tls"
	"io"
	"net"
	"runtime"
	"testing"
)

type closeWriter interface {
	CloseWrite() error
}

func TestBuildTCP(t *testing.T) {
	wc, co := Builder{TrackTraffic: true}.BuildWithObserver(new(net.TCPConn))
	if co == nil {
		t.Error("Expected a connection observer")
	}
	if _, ok := wc.(io.ReaderFrom); ok != (runtime.GOOS == "linux") {
		t.Error("ReaderFrom missmatch")
	}
	if _, ok := wc.(io.WriterTo); ok {
		t.Error("Unexpected WriterTo")
	}
	if _, ok := wc.(closeWriter); !ok {
		t.Error("Missing CloseWrite")
	}

	if ObserverFromConn(wc) != co {
		t.Error("ObserverFromConn mismatch")
	}
}

func TestBuildTLS(t *testing.T) {
	wc, co := Builder{TrackTraffic: true}.BuildWithObserver(new(tls.Conn))
	if co == nil {
		t.Error("Expected a connection observer")
	}
	if _, ok := wc.(io.ReaderFrom); ok {
		t.Error("Unexpected ReaderFrom")
	}
	if _, ok := wc.(io.WriterTo); ok {
		t.Error("Unexpected WriterTo")
	}
	if _, ok := wc.(closeWriter); !ok {
		t.Error("Missing CloseWrite")
	}

	if ObserverFromConn(wc) != co {
		t.Error("ObserverFromConn mismatch")
	}
}

func TestBuildOnClose(t *testing.T) {
	var closed bool
	wc, co := Builder{OnClose: func() { closed = true }}.BuildWithObserver(new(net.TCPConn))
	if co != nil {
		t.Error("Unexpected connection observer")
	}
	if _, ok := wc.(io.ReaderFrom); ok != (runtime.GOOS == "linux") {
		t.Error("ReaderFrom missmatch")
	}
	if _, ok := wc.(io.WriterTo); ok {
		t.Error("Unexpected WriterTo")
	}
	if _, ok := wc.(closeWriter); !ok {
		t.Error("Missing CloseWrite")
	}
	wc.Close()
	if !closed {
		t.Error("OnClose not called")
	}

	if ObserverFromConn(wc) != co {
		t.Error("ObserverFromConn mismatch")
	}
}
