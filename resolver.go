// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net"
	_ "unsafe" // for go:linkname
)

// lookupStaticHost looks up the addresses and the canonical name for the given host from /etc/hosts.
// It automatically updates the data cache.
//
//go:linkname lookupStaticHost net.lookupStaticHost
func lookupStaticHost(string) ([]string, string)

func isLocalhost(host string) bool {
	if host == "localhost" || host == "0.0.0.0" || host == "::" {
		return true
	}

	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}

	addrs, _ := lookupStaticHost(host)
	if len(addrs) > 0 {
		if ip := net.ParseIP(addrs[0]); ip != nil {
			return ip.IsLoopback()
		}
	}

	return false
}
