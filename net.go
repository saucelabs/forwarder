// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/ratelimit"
	"go.uber.org/multierr"
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
}

func DefaultDialConfig() *DialConfig {
	return &DialConfig{
		DialTimeout: 10 * time.Second,
		KeepAlive:   true,
	}
}

type Dialer struct {
	nd net.Dialer
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
		nd: nd,
	}
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.nd.DialContext(ctx, network, address)
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

type ListenerCallbacks interface {
	// OnAccept is called when a new connection is successfully accepted.
	OnAccept(net.Conn)

	// OnBindError is called when a listener fails to bind to an address.
	OnBindError(address string, err error)

	// OnTLSHandshakeError is called after a TLS handshake errors out.
	OnTLSHandshakeError(*tls.Conn, error)
}

// Listener is a multi-address listener with TLS support, rate limiting and callbacks.
// The Listener must successfully bind on Address, but it may fail to bind on OptionalAddresses.
type Listener struct {
	Address             string
	OptionalAddresses   []string
	Log                 log.Logger
	TLSConfig           *tls.Config
	TLSHandshakeTimeout time.Duration
	ReadLimit           int64
	WriteLimit          int64
	Callbacks           ListenerCallbacks

	listeners []net.Listener
	acceptCh  chan acceptResult
	wg        sync.WaitGroup
	closeCh   chan struct{}
	closeOnce sync.Once
}

type acceptResult struct {
	c   net.Conn
	err error
}

// Listen starts listening on the provided addresses.
// The method should be called only once.
func (l *Listener) Listen() error {
	l.acceptCh = make(chan acceptResult)
	l.closeCh = make(chan struct{})

	if err := l.listen(l.Address); err != nil {
		return err
	}

	// OptionalAddresses may fail to bind.
	for _, addr := range l.OptionalAddresses {
		if err := l.listen(addr); err != nil {
			l.Log.Errorf("failed to listen on %s: %v", addr, err)
			continue
		}
	}

	return nil
}

func (l *Listener) listen(addr string) error {
	ll, err := Listen("tcp", addr)
	if err != nil {
		if l.Callbacks != nil {
			l.Callbacks.OnBindError(addr, err)
		}
		return err
	}

	if rl, wl := l.ReadLimit, l.WriteLimit; rl > 0 || wl > 0 {
		// Notice that the ReadLimit stands for the read limit *from* a proxy, and the WriteLimit
		// stands for the write limit *to* a proxy, thus the ReadLimit is in fact
		// a txBandwidth and the WriteLimit is a rxBandwidth.
		ll = ratelimit.NewListener(ll, wl, rl)
	}

	l.listeners = append(l.listeners, ll)
	l.wg.Add(1)
	go l.acceptLoop(ll)

	return nil
}

func (l *Listener) acceptLoop(ll net.Listener) {
	defer l.wg.Done()
	for {
		c, err := ll.Accept()
		select {
		case l.acceptCh <- acceptResult{c, err}:
		case <-l.closeCh:
			if c != nil {
				if cerr := c.Close(); cerr != nil {
					l.Log.Errorf("failed to close connection: %v", cerr)
				}
			}
			return
		}
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	for {
		var (
			c   net.Conn
			err error
		)
		select {
		case <-l.closeCh:
			return nil, net.ErrClosed
		case res := <-l.acceptCh:
			c, err = res.c, res.err
		}
		if err != nil {
			return nil, err
		}

		if l.Callbacks != nil {
			l.Callbacks.OnAccept(c)
		}

		if l.TLSConfig == nil {
			return c, nil
		}

		tc, err := l.withTLS(c)
		if err != nil {
			l.Log.Errorf("failed to perform TLS handshake: %v", err)
			if cerr := tc.Close(); cerr != nil {
				l.Log.Errorf("failed to close connection: %v", cerr)
			}
			continue
		}

		return tc, nil
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
	if err != nil {
		if l.Callbacks != nil {
			l.Callbacks.OnTLSHandshakeError(tconn, err)
		}
	}

	return tconn, err
}

func (l *Listener) Addr() net.Addr {
	if len(l.listeners) == 0 {
		return &net.IPAddr{}
	}

	return l.listeners[0].Addr()
}

func (l *Listener) Addrs() []net.Addr {
	addrs := make([]net.Addr, 0, len(l.listeners))
	for _, ll := range l.listeners {
		addrs = append(addrs, ll.Addr())
	}
	return addrs
}

func (l *Listener) Close() error {
	l.closeOnce.Do(func() { close(l.closeCh) })

	var merr error
	for _, ll := range l.listeners {
		if err := ll.Close(); err != nil {
			merr = multierr.Append(merr, err)
		}
	}

	l.wg.Wait()

	return merr
}
