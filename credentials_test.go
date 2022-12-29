// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import (
	"net/url"
	"testing"

	"github.com/saucelabs/forwarder/log/stdlog"
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

			m, err := NewCredentialsMatcher(credentials, stdlog.Default())
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
