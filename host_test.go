// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"strings"
	"testing"
)

func TestParseHostPortUser(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "normal",
			input: "user:pass@foo:80",
		},
		{
			name:  "no user",
			input: ":pass@foo:80",
			err:   "username cannot be empty",
		},
		{
			name:  "empty",
			input: "",
			err:   "expected user[:password]@host:port",
		},
		{
			name:  "colon in password",
			input: "user:pass:pass@foo:80",
		},
		{
			name:  "@ in password",
			input: "user:p@ss@foo:80",
		},
		{
			name:  "@ in username",
			input: "user@:pass@foo:80",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			hpi, err := ParseHostPortUser(tc.input)
			if tc.err == "" {
				if err != nil {
					t.Fatalf("expected success, got %q", err)
				}
				pass, ok := hpi.Password()
				if ok {
					pass = ":" + pass
				}
				if hpi.Username()+pass+"@"+hpi.Host+":"+hpi.Port != tc.input {
					t.Errorf("expected %q, got %q", tc.input, hpi.String())
				}
			} else if !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("expected error to contain %q, got %q", tc.err, err)
			}
		})
	}
}
