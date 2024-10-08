// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package connfix

import (
	"bytes"
	"io"
	"net"
	"testing"
)

type testConn struct {
	net.Conn
}

var testAddr = &net.TCPAddr{}

func (tc testConn) RemoteAddr() net.Addr {
	return testAddr
}

func (tc testConn) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte("test"))
	return int64(n), err
}

func TestCombineTCPConn(t *testing.T) {
	tconn := new(net.TCPConn)
	if flags(tconn) == 0 {
		t.Fatal("flags(tconn) == 0")
	}

	t.Run("basic", func(t *testing.T) {
		conn := Combine(testConn{tconn}, tconn)
		if flags(conn) != flags(tconn) {
			t.Fatal("flags(conn) != flags(tconn)")
		}
		if conn.RemoteAddr() != testAddr {
			t.Fatal("conn.RemoteAddr() != testAddr")
		}
	})

	t.Run("overwrite", func(t *testing.T) {
		conn := Combine(testConn{tconn}, tconn)
		if flags(conn) != flags(tconn) {
			t.Fatal("flags(conn) != flags(tconn)")
		}
		var buf bytes.Buffer
		if _, err := conn.(io.WriterTo).WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		if buf.String() != "test" {
			t.Fatal("expected 'test', got", buf.String())
		}
	})

	t.Run("no overwrite", func(t *testing.T) {
		conn := Combine(testConn{tconn}, nil)
		if _, ok := conn.(io.WriterTo); ok {
			t.Fatal("expected no io.WriterTo")
		}
	})
}
