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

func NewListener(l net.Listener, rxBandwidth, txBandwidth int64) *Listener {
	var rxLimiter, txLimiter *rate.Limiter
	if rxBandwidth > 0 {
		rxLimiter = newRateLimiter(rxBandwidth)
	}
	if txBandwidth > 0 {
		txLimiter = newRateLimiter(txBandwidth)
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
