// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

var services = flag.String("services", "", "Comma-separated list of services to wait for")

func TestMain(m *testing.M) {
	flag.Parse()

	waitForDNS(strings.Split(*services, ","))

	os.Exit(m.Run())
}

func waitForDNS(services []string) {
	const d = 500 * time.Millisecond

	t := time.NewTicker(d)
	defer t.Stop()

	for {
		<-t.C

		pending := make([]string, 0)
		for _, s := range services {
			if !dnsReady(s) {
				pending = append(pending, s)
			}
		}
		if len(pending) == 0 {
			break
		}

		fmt.Fprintf(os.Stdout, "waiting for %s", strings.Join(pending, ", "))
	}
}

func dnsReady(name string) bool {
	a, err := net.LookupHost(name)
	if err != nil {
		return false
	}

	return len(a) != 0
}
