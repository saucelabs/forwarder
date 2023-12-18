// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"net"
	"syscall"
	"time"
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
