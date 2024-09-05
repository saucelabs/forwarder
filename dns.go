// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net/netip"
	"time"
)

type DNSConfig struct {
	Servers    []netip.AddrPort
	Timeout    time.Duration
	RoundRobin bool
}

func DefaultDNSConfig() *DNSConfig {
	return &DNSConfig{
		Timeout: 5 * time.Second,
	}
}
