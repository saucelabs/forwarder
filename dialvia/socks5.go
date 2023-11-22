// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dialvia

import (
	"context"
	"net"
	"net/url"

	"golang.org/x/net/proxy"
)

type SOCKS5ProxyDialer struct {
	dial     ContextDialerFunc
	proxyURL *url.URL
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

func (d *SOCKS5ProxyDialer) DialContext(_ context.Context, network, addr string) (net.Conn, error) {
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

	return sd.Dial(network, addr)
}
