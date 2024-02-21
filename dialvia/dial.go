// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dialvia

import (
	"context"
	"net"
)

// ContextDialerFunc is a function that implements Dialer and ContextDialer.
type ContextDialerFunc func(context context.Context, network, addr string) (net.Conn, error)

// Dial is needed to satisfy the proxy.Dialer interface.
// It is never called as proxy.ContextDialer is used instead if available.
func (f ContextDialerFunc) Dial(network, addr string) (net.Conn, error) {
	return f(context.Background(), network, addr)
}

func (f ContextDialerFunc) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return f(ctx, network, addr)
}
