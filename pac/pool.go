// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import (
	"net"
	"net/url"
	"sync"
)

type ProxyResolverPool struct {
	pool sync.Pool
}

func NewProxyResolverPool(cfg *ProxyResolverConfig, r *net.Resolver, opts ...Option) (*ProxyResolverPool, error) {
	if _, err := NewProxyResolver(cfg, r, opts...); err != nil {
		return nil, err
	}

	f := func() any {
		p, err := NewProxyResolver(cfg, r, opts...)
		if err != nil {
			panic(err)
		}
		return p
	}

	return &ProxyResolverPool{
		pool: sync.Pool{
			New: f,
		},
	}, nil
}

// FindProxyForURL calls FindProxyForURL or FindProxyForURLEx function in the PAC script with the alternate hostname.
// The hostname is optional, if empty it will be extracted from URL.
// This is to handle cases when the hostname is not a valid hostname, but a URL.
func (pool *ProxyResolverPool) FindProxyForURL(u *url.URL, hostname string) (p string, err error) {
	pr := pool.get()
	p, err = pr.FindProxyForURL(u, hostname)
	pool.pool.Put(pr)
	return
}

func (pool *ProxyResolverPool) get() *ProxyResolver {
	return pool.pool.Get().(*ProxyResolver) //nolint:forcetypeassert // we know it's a ProxyResolver
}
