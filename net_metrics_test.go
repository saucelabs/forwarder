// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"testing"
)

func TestAddr2HostLocalhost(t *testing.T) {
	addrs := []string{
		"localhost:80",
		"127.0.0.100:80",
		"[::1]:80",
		"[::]:52367",
	}

	for _, addr := range addrs {
		if host := addr2Host(addr); host != "localhost" {
			t.Fatalf("addr2Host(%q): got %q, want localhost", addr, host)
		}
	}
}
