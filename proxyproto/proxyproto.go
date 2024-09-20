// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package proxyproto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pires/go-proxyproto"
)

// Conn wraps a net.Conn and provides access to the proxy protocol header.
// If the header is not present or cannot be read within the timeout,
// the connection is closed.
type Conn struct {
	net.Conn

	readHeaderTimeout time.Duration
	isHeaderRead      atomic.Bool
	headerMu          sync.Mutex
	header            proxyproto.Header
	headerErr         error
}

func (c *Conn) NetConn() net.Conn {
	return c.Conn
}

func (c *Conn) LocalAddr() net.Addr {
	if err := c.readHeader(); err != nil {
		return c.Conn.LocalAddr()
	}

	if c.headerErr != nil || c.header.Command.IsLocal() {
		return c.Conn.LocalAddr()
	}

	return c.header.DestinationAddr
}

func (c *Conn) RemoteAddr() net.Addr {
	if err := c.readHeader(); err != nil {
		return c.Conn.RemoteAddr()
	}

	if c.headerErr != nil || c.header.Command.IsLocal() {
		return c.Conn.RemoteAddr()
	}

	return c.header.SourceAddr
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if err := c.readHeader(); err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if err := c.readHeader(); err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

func (c *Conn) Header() (proxyproto.Header, error) {
	return c.HeaderContext(context.Background())
}

func (c *Conn) HeaderContext(ctx context.Context) (proxyproto.Header, error) {
	if err := c.readHeaderContext(ctx); err != nil {
		return proxyproto.Header{}, err
	}
	return c.header, nil
}

func (c *Conn) readHeader() error {
	return c.readHeaderContext(context.Background())
}

func (c *Conn) readHeaderContext(ctx context.Context) error {
	if c.isHeaderRead.Load() {
		return c.headerErr
	}

	c.headerMu.Lock()
	defer c.headerMu.Unlock()

	if c.isHeaderRead.Load() {
		return c.headerErr
	}

	t0 := time.Now()
	if c.readHeaderTimeout > 0 {
		if d, ok := ctx.Deadline(); !ok || d.Sub(t0) > c.readHeaderTimeout {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.readHeaderTimeout)
			defer cancel()
		}
	}

	type result struct {
		header *proxyproto.Header
		err    error
	}
	resCh := make(chan result)

	go func() {
		// For v1 the header length is at most 108 bytes.
		// For v2 the header length is at most 52 bytes plus the length of the TLVs.
		// We use 256 bytes to be safe.
		const bufSize = 256
		// Use a byteReader to read only one byte at a time,
		// so we can read the header without consuming more bytes than needed.
		// On success, the reader must be empty.
		// Otherwise, the connection is closed on timeout or never read on error.
		br := bufio.NewReaderSize(byteReader{c.Conn}, bufSize)

		var r result
		r.header, r.err = proxyproto.Read(br)

		if r.err == nil && br.Buffered() > 0 {
			panic("proxy protocol header read: unexpected data after header")
		}

		resCh <- r
	}()

	select {
	case <-ctx.Done():
		c.Conn.Close()
		c.headerErr = fmt.Errorf("proxy protocol header read timeout: %w", ctx.Err())
	case r := <-resCh:
		c.header = *r.header
		c.headerErr = r.err
	}

	c.isHeaderRead.Store(true)

	return c.headerErr
}

type Listener struct {
	net.Listener
	ReadHeaderTimeout time.Duration
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{
		Conn:              c,
		readHeaderTimeout: l.ReadHeaderTimeout,
	}, nil
}

type byteReader struct {
	r io.Reader
}

func (r byteReader) Read(p []byte) (int, error) {
	return r.r.Read(p[:1])
}
