// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net/url"
	"testing"

	"github.com/saucelabs/forwarder/log/slog"
)

func TestUserInfoMatcherMatch(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		hostport string
		expected *url.Userinfo
	}{
		{
			name:     "Matcher is empty",
			hostport: "abc:80",
		},
		{
			name:     "Matches hostport",
			input:    []string{"user:pass@abc:80", "foo:pass@*:80", "bar:pass@abc:0", "baz:pass@*:0"},
			hostport: "abc:80",
			expected: url.UserPassword("user", "pass"),
		},
		{
			name:     "Matches host wildcard",
			input:    []string{"user:pass@abc:80", "foo:pass@*:80", "bar:pass@abc:0", "baz:pass@*:0"},
			hostport: "xxx:80",
			expected: url.UserPassword("foo", "pass"),
		},
		{
			name:     "Matches port wildcard",
			input:    []string{"user:pass@abc:80", "foo:pass@*:80", "bar:pass@abc:0", "baz:pass@*:0"},
			hostport: "abc:90",
			expected: url.UserPassword("bar", "pass"),
		},
		{
			name:     "Matches global wildcard",
			input:    []string{"user:pass@abc:80", "foo:pass@*:80", "bar:pass@abc:0", "baz:pass@*:0"},
			hostport: "xxx:443",
			expected: url.UserPassword("baz", "pass"),
		},
		{
			name:     "Matches port '*'",
			input:    []string{"user:pass@abc:*"},
			hostport: "abc:80",
			expected: url.UserPassword("user", "pass"),
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			var (
				credentials = make([]*HostPortUser, len(tc.input))
				err         error
			)
			for i := range tc.input {
				credentials[i], err = ParseHostPortUser(tc.input[i])
				if err != nil {
					t.Fatal(err)
				}
			}

			m, err := NewCredentialsMatcher(credentials, slog.Default())
			if err != nil {
				t.Fatal(err)
			}
			u := m.Match(tc.hostport)

			if u.String() != tc.expected.String() {
				t.Fatalf("expected %s, got %s", tc.expected, u)
			}
		})
	}
}
