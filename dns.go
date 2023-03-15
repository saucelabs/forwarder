// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/saucelabs/forwarder/log"
)

type DNSConfig struct {
	// Servers is a list of DNS servers to use.
	// If provided with multiple servers, the first one is used as primary server, the rest are used as a fallback.
	Servers []netip.AddrPort
	// Timeout is the timeout for DNS queries.
	Timeout time.Duration
}

func DefaultDNSConfig() *DNSConfig {
	return &DNSConfig{
		Timeout: 5 * time.Second,
	}
}

func (c *DNSConfig) Validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("no DNS server configured")
	}
	for i, u := range c.Servers {
		if err := validateDNSAddress(u); err != nil {
			return fmt.Errorf("servers[%d]: %w", i, err)
		}
	}
	return nil
}

type resolver struct {
	resolver net.Resolver
	dialer   net.Dialer
	servers  []netip.AddrPort
	log      log.Logger
}

func NewResolver(cfg *DNSConfig, log log.Logger) (*net.Resolver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log.Infof("using DNS servers: %s", cfg.Servers)

	r := new(resolver)
	*r = resolver{
		resolver: net.Resolver{
			PreferGo: true,
			Dial:     r.dialDNS,
		},
		dialer: net.Dialer{
			Timeout:  cfg.Timeout,
			Resolver: nopResolver(),
		},
		servers: cfg.Servers,
		log:     log,
	}

	return &r.resolver, nil
}

func nopResolver() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, fmt.Errorf("no DNS resolver configured")
		},
	}
}

func (r *resolver) dialDNS(ctx context.Context, network, address string) (net.Conn, error) {
	for _, u := range r.servers {
		r.log.Debugf("dial DNS server %s instead of %s", u, address)
		conn, err := r.dialer.DialContext(ctx, network, u.String())
		if err != nil {
			r.log.Errorf("failed to dial DNS server %s: %v", u, err)
			continue
		}
		return conn, nil
	}

	return nil, fmt.Errorf("failed to dial DNS server")
}
