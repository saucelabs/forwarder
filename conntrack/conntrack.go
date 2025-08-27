// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package conntrack

import (
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/saucelabs/connfu"
	"github.com/saucelabs/forwarder/utils/reflectx"
)

// Observer allows to observe the number of bytes read and written from a connection.
type Observer struct {
	rx atomic.Uint64
	tx atomic.Uint64
}

// Rx returns the number of bytes read from the connection.
// It requires TrackTraffic to be set to true, otherwise it returns 0.
func (o *Observer) Rx() uint64 {
	return o.rx.Load()
}

// Tx returns the number of bytes written to the connection.
// It requires TrackTraffic to be set to true, otherwise it returns 0.
func (o *Observer) Tx() uint64 {
	return o.tx.Load()
}

func (o *Observer) addRx(n uint64) {
	o.rx.Add(n)
}

func (o *Observer) addTx(n uint64) {
	o.tx.Add(n)
}

type closeConn struct {
	net.Conn
	l closeListener // this is a field to avoid ambiguous selector error on Close method
}

func (c *closeConn) Close() error {
	return c.l.Close()
}

type closeListener struct {
	close   func() error
	once    sync.Once
	onClose func()
}

func (c *closeListener) Close() error {
	err := c.close()
	c.once.Do(c.onClose)
	return err
}

// conn is a net.Conn that tracks the number of bytes read and written.
// It needs to be configured before first use by setting TrackTraffic and onClose if needed.
type conn struct {
	net.Conn
	o Observer
}

func (c *conn) Read(p []byte) (n int, err error) {
	n, err = c.Conn.Read(p)
	c.o.addRx(uint64(n)) //nolint:gosec // n is never negative.
	return
}

func (c *conn) Write(p []byte) (n int, err error) {
	n, err = c.Conn.Write(p)
	c.o.addTx(uint64(n)) //nolint:gosec // n is never negative.
	return
}

func (c *conn) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = c.Conn.(io.ReaderFrom).ReadFrom(r) //nolint:forcetypeassert // It is checked before.
	c.o.addTx(uint64(n))                        //nolint:gosec // n is never negative.
	return
}

func (c *conn) Observer() *Observer {
	return &c.o
}

type Builder struct {
	// TrackTraffic enables counting of bytes read and written by the connection.
	// Use Rx and Tx to get the number of bytes read and written.
	TrackTraffic bool

	// OnClose is called after the underlying connection is closed and before the Close method returns.
	// OnClose is called at most once.
	OnClose func()
}

func (b Builder) Build(c net.Conn) net.Conn {
	wc, _ := b.BuildWithObserver(c)
	return wc
}

func (b Builder) BuildWithObserver(c net.Conn) (net.Conn, *Observer) {
	var (
		wc net.Conn
		co *Observer
	)

	if b.TrackTraffic {
		if b.OnClose != nil {
			cc := &struct {
				conn
				closeListener
			}{
				conn: conn{Conn: c},
				closeListener: closeListener{
					close:   c.Close,
					onClose: b.OnClose,
				},
			}
			wc = cc
			co = &cc.conn.o
		} else {
			cc := &conn{
				Conn: c,
			}
			wc = cc
			co = &cc.o
		}
	} else {
		if b.OnClose == nil {
			wc = c
		} else {
			wc = &closeConn{
				Conn: c,
				l: closeListener{
					close:   c.Close,
					onClose: b.OnClose,
				},
			}
		}
	}

	return connfu.Combine(wc, c), co
}

func ObserverFromConn(conn net.Conn) *Observer {
	type ifce interface {
		Observer() *Observer
	}

	if o, ok := conn.(ifce); ok {
		return o.Observer()
	}

	v, ok := reflectx.LookupImpl[ifce](reflect.ValueOf(conn))
	if !ok {
		return nil
	}

	return v.Observer()
}
