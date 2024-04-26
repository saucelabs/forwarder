// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dialvia

import (
	"context"
	"net"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

type SOCKS5ProxyDialer struct {
	dial     ContextDialerFunc
	proxyURL *url.URL

	Timeout time.Duration
}

func SOCKS5Proxy(dial ContextDialerFunc, proxyURL *url.URL) *SOCKS5ProxyDialer {
	if dial == nil {
		panic("dial is required")
	}
	if proxyURL == nil {
		panic("proxy URL is required")
	}
	if proxyURL.Scheme != "socks5" {
		panic("proxy URL scheme must be socks5")
	}

	return &SOCKS5ProxyDialer{
		dial:     dial,
		proxyURL: proxyURL,
	}
}

func (d *SOCKS5ProxyDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if d.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.Timeout)
		defer cancel()
	}

	u := d.proxyURL.User
	var auth *proxy.Auth
	if u != nil {
		auth = new(proxy.Auth)
		auth.User = u.Username()
		if p, ok := u.Password(); ok {
			auth.Password = p
		}
	}

	proxyHost := d.proxyURL.Hostname()
	proxyPort := d.proxyURL.Port()
	if proxyPort == "" {
		proxyPort = "1080"
	}
	proxyAddr := net.JoinHostPort(proxyHost, proxyPort)

	sd, err := proxy.SOCKS5("tcp", proxyAddr, auth, d.dial)
	if err != nil {
		return nil, err
	}

	sdctx := sd.(contextDialer) //nolint:forcetypeassert // I want it to panic if it's not a ContextDialerFunc.
	return sdctx.DialContext(ctx, network, addr)
}

type contextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}
