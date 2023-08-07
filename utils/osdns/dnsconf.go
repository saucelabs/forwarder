// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at https://github.com/golang/go/blob/-/LICENSE.

package osdns

import (
	_ "net" // for go:linkname
	"net/netip"
	"time"
	_ "unsafe" // for go:linkname
)

// dnsConfig is 1:1 copy of net.dnsConfig struct.
// Source: https://github.com/golang/go/blob/-/src/net/dnsconfig.go
// Do not modify.
//
//nolint:unused // used by go:linkname
type dnsConfig struct {
	servers       []string      // server addresses (in host:port form) to use
	search        []string      // rooted suffixes to append to local name
	ndots         int           // number of dots in name to trigger absolute lookup
	timeout       time.Duration // wait before giving up on a query, including retries
	attempts      int           // lost packets before giving up on server
	rotate        bool          // round robin among servers
	unknownOpt    bool          // anything unknown was encountered
	lookup        []string      // OpenBSD top-level database "lookup" order
	err           error         // any error that occurs during open of resolv.conf
	mtime         time.Time     // time of resolv.conf modification
	soffset       uint32        // used by serverOffset
	singleRequest bool          // use sequential A and AAAA queries instead of parallel queries
	useTCP        bool          // force usage of TCP for DNS resolutions
	trustAD       bool          // add AD flag to queries
	noReload      bool          // do not check for config file updates
}

//go:linkname getSystemDNSConfig net.getSystemDNSConfig
func getSystemDNSConfig() *dnsConfig

type Config struct {
	Servers    []netip.AddrPort
	Timeout    time.Duration
	RoundRobin bool
}

func DefaultConfig() *Config {
	return &Config{
		Timeout: 5 * time.Second,
	}
}
