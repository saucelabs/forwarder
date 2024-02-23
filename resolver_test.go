// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"testing"
)

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		host      string
		localhost bool
	}{
		{"127.0.0.1", true},
		{"127.10.20.30", true},
		{"localhost", true},

		{"notlocalhost", false},
		{"broadcasthost", false},

		{"::1", true},

		{"::10", false},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
	}

	for i := range tests {
		tc := tests[i]
		if lh := isLocalhost(tc.host); lh != tc.localhost {
			t.Errorf("isLocalhost(%q) = %v; want %v", tc.host, lh, tc.localhost)
		}
	}
}
