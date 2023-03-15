// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/log/stdlog"
)

func TestResolverLookupHost(t *testing.T) {
	if _, ok := os.LookupEnv("CI"); ok {
		t.Skip("skipping test in CI environment")
	}

	dnsAddr, err := netip.ParseAddrPort("1.1.1.1:53")
	if err != nil {
		t.Fatal(err)
	}

	c := &DNSConfig{
		Servers: []netip.AddrPort{dnsAddr},
		Timeout: 5 * time.Second,
	}
	r, err := NewResolver(c, stdlog.Default())
	if err != nil {
		t.Fatal(err)
	}

	addr, err := r.LookupHost(context.Background(), "google.com")
	if err != nil {
		t.Errorf("LookupHost failed: %v", err)
	}
	if len(addr) == 0 {
		t.Errorf("LookupHost returned no address")
	}
}
