// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ratelimit

import (
	"net"

	"golang.org/x/time/rate"
)

type Listener struct {
	net.Listener
	rxLimiter *rate.Limiter
	txLimiter *rate.Limiter
}

// NewListener creates a new rate-limited listener.
// The readLimit and writeLimit should be seen from the perspective of a peer that opens a connection to this listener.
// How much they can read and write, respectively.
// Limits are in bytes per second.
func NewListener(l net.Listener, readLimit, writeLimit int64) *Listener {
	// Notice that the readLimit should be seen from the perspective of a peer that opens a connection to this listener.
	// Thus, the readLimit is in fact a txBandwidth - How much data can be sent to the peer that opened the connection,
	// controls how much data they can *read*.
	// The same goes for writeLimit.
	var rxLimiter, txLimiter *rate.Limiter
	if readLimit > 0 {
		txLimiter = newRateLimiter(readLimit)
	}
	if writeLimit > 0 {
		rxLimiter = newRateLimiter(writeLimit)
	}

	return &Listener{
		Listener:  l,
		rxLimiter: rxLimiter,
		txLimiter: txLimiter,
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &Conn{
		Conn:      c,
		rxLimiter: l.rxLimiter,
		txLimiter: l.txLimiter,
	}, nil
}
