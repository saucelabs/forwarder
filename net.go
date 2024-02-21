// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/ratelimit"
)

type DialConfig struct {
	// DialTimeout is the maximum amount of time a dial will wait for
	// connect to complete.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	DialTimeout time.Duration

	// KeepAlive enables TCP keep-alive probes for an active network connection.
	// The keep-alive probes are sent with OS specific intervals.
	KeepAlive bool

	PromConfig
}

func DefaultDialConfig() *DialConfig {
	return &DialConfig{
		DialTimeout: 10 * time.Second,
		KeepAlive:   true,
	}
}

type Dialer struct {
	nd      net.Dialer
	metrics *dialerMetrics
}

func NewDialer(cfg *DialConfig) *Dialer {
	nd := net.Dialer{
		Timeout:   cfg.DialTimeout,
		KeepAlive: -1,
		Resolver: &net.Resolver{
			PreferGo: true,
		},
	}

	if cfg.KeepAlive {
		nd.Control = func(network, address string, c syscall.RawConn) error {
			return c.Control(enableTCPKeepAlive)
		}
	}

	return &Dialer{
		nd:      nd,
		metrics: newDialerMetrics(cfg.PromRegistry, cfg.PromNamespace),
	}
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	conn, err := d.nd.DialContext(ctx, network, address)
	if err != nil {
		d.metrics.error(address)
		return nil, err
	}

	d.metrics.dial(address)

	return &TrackedConn{
		Conn: conn,
		OnClose: func() {
			d.metrics.close(address)
		},
	}, nil
}

func defaultListenConfig() *net.ListenConfig {
	return &net.ListenConfig{
		KeepAlive: -1,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(enableTCPKeepAlive)
		},
	}
}

// Listen creates a listener for the provided network and address and configures OS-specific keep-alive parameters.
// See net.Listen for more information.
func Listen(network, address string) (net.Listener, error) {
	// The context cancellation does not close the listener.
	// I asked about it here: https://groups.google.com/g/golang-nuts/c/Q1I7Viz9AJc
	return defaultListenConfig().Listen(context.Background(), network, address)
}

type Listener struct {
	Address             string
	Log                 log.Logger
	TLSConfig           *tls.Config
	TLSHandshakeTimeout time.Duration
	ReadLimit           int64
	WriteLimit          int64
	PromConfig

	listener net.Listener
	metrics  *listenerMetrics
}

func (l *Listener) Listen() error {
	if l.listener != nil {
		return fmt.Errorf("already listening on %s", l.Address)
	}

	ll, err := Listen("tcp", l.Address)
	if err != nil {
		return err
	}

	if rl, wl := l.ReadLimit, l.WriteLimit; rl > 0 || wl > 0 {
		// Notice that the ReadLimit stands for the read limit *from* a proxy, and the WriteLimit
		// stands for the write limit *to* a proxy, thus the ReadLimit is in fact
		// a txBandwidth and the WriteLimit is a rxBandwidth.
		ll = ratelimit.NewListener(ll, wl, rl)
	}

	l.listener = ll
	l.metrics = newListenerMetrics(l.PromRegistry, l.PromNamespace)

	return nil
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			l.metrics.error()
			return nil, err
		}

		if l.TLSConfig == nil {
			l.metrics.accept()
			return &TrackedConn{
				Conn:    conn,
				OnClose: l.metrics.close,
			}, nil
		}

		tr := &TrackedConn{
			Conn: conn,
		}
		tconn, err := l.withTLS(tr)
		if err != nil {
			l.Log.Errorf("TLS handshake failed: %v", err)
			if cerr := tconn.Close(); cerr != nil {
				l.Log.Errorf("error while closing TLS connection after failed handshake: %v", cerr)
			}
			l.metrics.tlsError()

			continue
		}

		l.metrics.accept()
		tr.OnClose = l.metrics.close
		return tconn, nil
	}
}

func (l *Listener) withTLS(conn net.Conn) (*tls.Conn, error) {
	tconn := tls.Server(conn, l.TLSConfig)

	var err error
	if l.TLSHandshakeTimeout <= 0 {
		err = tconn.Handshake()
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), l.TLSHandshakeTimeout)
		err = tconn.HandshakeContext(ctx)
		cancel()
	}

	return tconn, err
}

func (l *Listener) Addr() net.Addr {
	if l.listener == nil {
		return nil
	}
	return l.listener.Addr()
}

func (l *Listener) Close() error {
	if l.listener == nil {
		return nil
	}
	return l.listener.Close()
}

// TrackedConn is a net.Conn that tracks the number of bytes read and written.
// It needs to be configured before first use by setting TrackTraffic and OnClose if needed.
type TrackedConn struct {
	net.Conn

	// TrackTraffic enables counting of bytes read and written by the connection.
	// Use Rx and Tx to get the number of bytes read and written.
	TrackTraffic bool

	// OnClose is called after the underlying connection is closed and before the Close method returns.
	// OnClose is called at most once.
	OnClose func()

	rx   atomic.Uint64
	tx   atomic.Uint64
	once sync.Once
}

func (c *TrackedConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if c.TrackTraffic {
		c.rx.Add(uint64(n))
	}
	return n, err
}

func (c *TrackedConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	if c.TrackTraffic {
		c.tx.Add(uint64(n))
	}
	return n, err
}

// Rx returns the number of bytes read from the connection.
// It requires TrackTraffic to be set to true, otherwise it returns 0.
func (c *TrackedConn) Rx() uint64 {
	return c.rx.Load()
}

// Tx returns the number of bytes written to the connection.
// It requires TrackTraffic to be set to true, otherwise it returns 0.
func (c *TrackedConn) Tx() uint64 {
	return c.tx.Load()
}

func (c *TrackedConn) Close() error {
	err := c.Conn.Close()
	if c.OnClose != nil {
		c.once.Do(c.OnClose)
	}
	return err
}
