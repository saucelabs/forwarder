// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package proxyproto

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/saucelabs/connfu"
)

// Conn wraps a net.Conn and provides access to the proxy protocol header.
// If the header is not present or cannot be read within the timeout,
// the connection is closed.
type Conn struct {
	net.Conn

	readHeaderTimeout time.Duration
	isHeaderRead      atomic.Bool
	headerMu          sync.Mutex
	header            Header
	headerErr         error
}

func (c *Conn) LocalAddr() net.Addr {
	if err := c.readHeader(); err != nil {
		return c.Conn.LocalAddr()
	}

	if c.headerErr != nil || c.header.IsLocal {
		return c.Conn.LocalAddr()
	}

	return c.header.Destination
}

func (c *Conn) RemoteAddr() net.Addr {
	if err := c.readHeader(); err != nil {
		return c.Conn.RemoteAddr()
	}

	if c.headerErr != nil || c.header.IsLocal {
		return c.Conn.RemoteAddr()
	}

	return c.header.Source
}

func (c *Conn) Read(b []byte) (n int, err error) {
	if err := c.readHeader(); err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

func (c *Conn) ReadFrom(r io.Reader) (n int64, err error) {
	if err := c.readHeader(); err != nil {
		return 0, err
	}
	return c.Conn.(io.ReaderFrom).ReadFrom(r) //nolint:forcetypeassert // handled by connfu.Combine
}

func (c *Conn) Write(b []byte) (n int, err error) {
	if err := c.readHeader(); err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

func (c *Conn) WriteTo(w io.Writer) (n int64, err error) {
	if err := c.readHeader(); err != nil {
		return 0, err
	}
	return c.Conn.(io.WriterTo).WriteTo(w) //nolint:forcetypeassert // handled by connfu.Combine
}

func (c *Conn) Header() (Header, error) {
	return c.HeaderContext(context.Background())
}

func (c *Conn) HeaderContext(ctx context.Context) (Header, error) {
	if err := c.readHeaderContext(ctx); err != nil {
		return Header{}, err
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
		header *Header
		err    error
	}
	resCh := make(chan result, 1)

	go func() {
		var r result
		r.header, r.err = ReadHeader(c.Conn)
		resCh <- r
	}()

	select {
	case <-ctx.Done():
		c.Conn.Close()
		c.headerErr = fmt.Errorf("proxy protocol header read timeout: %w", ctx.Err())
	case r := <-resCh:
		if r.header != nil {
			c.header = *r.header
		}
		c.headerErr = r.err
	}

	c.isHeaderRead.Store(true)

	return c.headerErr
}

type Listener struct {
	net.Listener
	ReadHeaderTimeout time.Duration
	TestingSkipConnfu bool
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	pc := &Conn{
		Conn:              c,
		readHeaderTimeout: l.ReadHeaderTimeout,
	}

	if l.TestingSkipConnfu {
		c = pc
	} else {
		c = connfu.Combine(pc, c)
	}

	return c, nil
}
