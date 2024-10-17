// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package proxyproto

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/net/nettest"
)

//go:embed testdata/v2.bin
var v2Header []byte

func makePipeV1() (c1, c2 net.Conn, stop func(), err error) {
	return makePipe([]byte("PROXY TCP4 1.1.1.1 2.2.2.2 1000 2000\r\n"))
}

func makePipeV2() (c1, c2 net.Conn, stop func(), err error) {
	return makePipe(v2Header)
}

func makePipe(header []byte) (c1, c2 net.Conn, stop func(), err error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, nil, nil, err
	}
	l = &Listener{
		Listener:          l,
		TestingSkipConnfu: true,
	}

	// Start a connection between two endpoints.
	var err1, err2 error
	done := make(chan bool)
	go func() {
		c2, err2 = l.Accept()
		close(done)
	}()
	c1, err1 = net.Dial(l.Addr().Network(), l.Addr().String())
	if err1 == nil {
		_, err1 = c1.Write(header)
	}
	<-done

	stop = func() {
		if err1 == nil {
			c1.Close()
		}
		if err2 == nil {
			c2.Close()
		}
		l.Close()
	}

	switch {
	case err1 != nil:
		stop()
		return nil, nil, nil, err1
	case err2 != nil:
		stop()
		return nil, nil, nil, err2
	default:
		return c1, c2, stop, nil
	}
}

func TestTestConn(t *testing.T) {
	t.Parallel()

	t.Run("v1", func(t *testing.T) {
		t.Parallel()
		nettest.TestConn(t, makePipeV1)
	})

	t.Run("v2", func(t *testing.T) {
		t.Parallel()
		nettest.TestConn(t, makePipeV2)
	})
}

func TestConnHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version  int
		makePipe func() (net.Conn, net.Conn, func(), error)
	}{
		{1, makePipeV1},
		{2, makePipeV2},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(fmt.Sprintf("v%d", tc.version), func(t *testing.T) {
			t.Parallel()

			_, c2, stop, err := tc.makePipe()
			if err != nil {
				t.Fatal(err)
			}
			defer stop()

			e := Header{
				Source:      &net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 1000},
				Destination: &net.TCPAddr{IP: net.ParseIP("2.2.2.2"), Port: 2000},
				IsLocal:     false,
				Version:     tc.version,
			}

			h, err := c2.(*Conn).Header()
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(e, h); diff != "" {
				t.Errorf("unexpected header (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConnReadHeaderTimeout(t *testing.T) {
	t.Parallel()

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	l = &Listener{
		Listener:          l,
		ReadHeaderTimeout: 50 * time.Millisecond,
	}

	connCh := make(chan net.Conn)

	go func() {
		conn, err := net.Dial("tcp", l.Addr().String())
		if err != nil {
			t.Error(err)
			return
		}
		if _, err := conn.Write([]byte("PROXY TCP4")); err != nil {
			t.Error(err)
			return
		}

		connCh <- conn
	}()

	conn, err := l.Accept()
	if err != nil {
		t.Error(err)
		return
	}

	if _, err = io.ReadAll(conn); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	if err := conn.Close(); !strings.Contains(err.Error(), "use of closed network connection") {
		t.Errorf("expected use of closed network connection, got %v", err)
	}

	(<-connCh).Close()
}
