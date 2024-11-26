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
	"slices"
	"time"

	"github.com/saucelabs/forwarder/conntrack"
	"github.com/saucelabs/forwarder/proxyproto"
	"github.com/saucelabs/forwarder/ratelimit"
)

type DialRedirectFunc func(network, address string) (targetNetwork, targetAddress string)

func DialRedirectFromHostPortPairs(subs []HostPortPair) DialRedirectFunc {
	subs = slices.Clone(subs)
	return func(network, address string) (string, string) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return network, address
		}

		for _, s := range subs {
			if (s.Src.Host == "" || s.Src.Host == host) && (s.Src.Port == "" || s.Src.Port == port) { //nolint:gocritic // nestingReduce: invert if cond, replace body with `continue`, move old body after the statement
				h := s.Dst.Host
				if h == "" {
					h = host
				}
				p := s.Dst.Port
				if p == "" {
					p = port
				}
				return network, net.JoinHostPort(h, p)
			}
		}

		return network, address
	}
}

// defaultKeepAliveConfig returns configuration that enables keep-alive pings with OS-specific parameters.
func defaultKeepAliveConfig() net.KeepAliveConfig {
	return net.KeepAliveConfig{
		Enable:   true,
		Idle:     -1,
		Interval: -1,
		Count:    -1,
	}
}

type DialConfig struct {
	// DialTimeout is the maximum amount of time a dial will wait for
	// connect to complete.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	DialTimeout time.Duration

	// KeepAliveConfig contains TCP keep-alive options.
	KeepAliveConfig net.KeepAliveConfig

	// RedirectFunc can be optionally set to redirect the connection to a different address.
	RedirectFunc DialRedirectFunc

	PromConfig
}

func DefaultDialConfig() *DialConfig {
	return &DialConfig{
		DialTimeout:     25 * time.Second,
		KeepAliveConfig: defaultKeepAliveConfig(),
	}
}

type Dialer struct {
	nd      net.Dialer
	rd      DialRedirectFunc
	metrics *dialerMetrics
}

func NewDialer(cfg *DialConfig) *Dialer {
	nd := net.Dialer{
		Timeout:         cfg.DialTimeout,
		KeepAlive:       -1,
		KeepAliveConfig: cfg.KeepAliveConfig,
		Resolver: &net.Resolver{
			PreferGo: true,
		},
	}

	return &Dialer{
		nd:      nd,
		rd:      cfg.RedirectFunc,
		metrics: newDialerMetrics(cfg.PromRegistry, cfg.PromNamespace),
	}
}

// DialConnTrack specifies the connection tracking mode for connections dialed by Dialer.
type DialConnTrack uint8

const (
	DialConnTrackDefault DialConnTrack = iota
	DialConnTrackDisabled
	DialConnTrackTraffic
)

type dialConnTrackKey struct{}

// WithDialConnTrack sets the connection tracking mode for connections dialed by Dialer.
func WithDialConnTrack(ctx context.Context, track DialConnTrack) context.Context {
	return context.WithValue(ctx, dialConnTrackKey{}, track)
}

// DialContext dials the provided network and address and configures OS-specific keep-alive parameters.
// It tracks dialed and closed connections by default, the behavior can be changed with WithDialConnTrack.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var dct DialConnTrack
	if v, ok := ctx.Value(dialConnTrackKey{}).(DialConnTrack); ok {
		dct = v
	}

	if d.rd != nil {
		network, address = d.rd(network, address)
	}
	conn, err := d.nd.DialContext(ctx, network, address)

	if dct == DialConnTrackDisabled {
		return conn, err
	}

	if err != nil {
		d.metrics.error(address)
		return nil, err
	}

	d.metrics.dial(address)

	return conntrack.Builder{
		TrackTraffic: dct == DialConnTrackTraffic,
		OnClose: func() {
			d.metrics.close(address)
		},
	}.Build(conn), nil
}

type ProxyProtocolConfig struct {
	ReadHeaderTimeout time.Duration
}

func DefaultProxyProtocolConfig() *ProxyProtocolConfig {
	return &ProxyProtocolConfig{
		ReadHeaderTimeout: 5 * time.Second,
	}
}

type ListenerConfig struct {
	Address             string
	KeepAliveConfig     net.KeepAliveConfig
	ProxyProtocolConfig *ProxyProtocolConfig
	ReadLimit           SizeSuffix
	WriteLimit          SizeSuffix
	TrackTraffic        bool
}

func DefaultListenerConfig(addr string) *ListenerConfig {
	return &ListenerConfig{
		Address:         addr,
		KeepAliveConfig: defaultKeepAliveConfig(),
	}
}

type NamedListenerConfig struct {
	Name string
	ListenerConfig
}

// MultiListener is a builder for multiple listeners sharing the same prometheus configuration.
// The listener name is added as a label to the metrics.
type MultiListener struct {
	ListenerConfigs []NamedListenerConfig
	TLSConfig       func(NamedListenerConfig) *tls.Config
	PromConfig
}

func (ml MultiListener) Listen() (_ []net.Listener, ferr error) {
	listeners := make([]net.Listener, 0, len(ml.ListenerConfigs))
	defer func() {
		if ferr != nil {
			for _, l := range listeners {
				l.Close()
			}
		}
	}()

	mf := newListenerMetricsWithNameFunc(ml.PromRegistry, ml.PromNamespace)

	for _, lc := range ml.ListenerConfigs {
		l := new(Listener)
		l.ListenerConfig = lc.ListenerConfig
		if ml.TLSConfig != nil {
			l.TLSConfig = ml.TLSConfig(lc)
		}
		l.metrics = mf(lc.Name)
		if err := l.Listen(); err != nil {
			return nil, err
		}
		listeners = append(listeners, l)
	}

	return listeners, nil
}

type Listener struct {
	ListenerConfig
	TLSConfig *tls.Config
	PromConfig

	listener net.Listener
	metrics  *listenerMetrics
}

func (l *Listener) Listen() error {
	if l.listener != nil {
		return fmt.Errorf("already listening on %s", l.Address)
	}

	ll, err := l.listen()
	if err != nil {
		return err
	}

	if l.ProxyProtocolConfig != nil {
		ll = &proxyproto.Listener{
			Listener:          ll,
			ReadHeaderTimeout: l.ProxyProtocolConfig.ReadHeaderTimeout,
		}
	}

	if rl, wl := l.ReadLimit, l.WriteLimit; rl > 0 || wl > 0 {
		ll = ratelimit.NewListener(ll, int64(rl), int64(wl))
	}

	l.listener = ll

	if l.metrics == nil {
		l.metrics = newListenerMetrics(l.PromRegistry, l.PromNamespace)
	}

	return nil
}

func (l *Listener) listen() (net.Listener, error) {
	lc := &net.ListenConfig{
		KeepAlive:       -1,
		KeepAliveConfig: l.ListenerConfig.KeepAliveConfig,
	}
	// The context cancellation does not close the listener.
	// I asked about it here: https://groups.google.com/g/golang-nuts/c/Q1I7Viz9AJc
	return lc.Listen(context.Background(), "tcp", l.Address)
}

// Accept returns tls.Conn if TLSConfig is set, as martian expects it to be on top.
// Otherwise, it returns forwarder.TrackedConn.
func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		l.metrics.error()
		return nil, err
	}

	l.metrics.accept()
	conn = conntrack.Builder{
		TrackTraffic: l.TrackTraffic,
		OnClose:      l.metrics.close,
	}.Build(conn)

	if l.TLSConfig != nil {
		conn = tls.Server(conn, l.TLSConfig)
	}

	return conn, nil
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
