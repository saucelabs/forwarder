// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ratelimit

import (
	"context"
	"net"

	"golang.org/x/time/rate"
)

type Conn struct {
	net.Conn
	rxLimiter *rate.Limiter
	txLimiter *rate.Limiter
}

var waitContext = context.Background()

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if n > 0 && c.rxLimiter != nil {
		c.rxLimiter.WaitN(waitContext, n)
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if n > 0 && c.txLimiter != nil {
		c.txLimiter.WaitN(waitContext, n)
	}
	return
}
