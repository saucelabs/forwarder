// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"net/url"
	"strings"
	"testing"
)

func TestNewUserInfoMatcherErrors(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		err   string
	}{
		{
			name:  "Empty user",
			input: []string{":pass@abc"},
			err:   "missing username",
		},
		{
			name:  "Empty password",
			input: []string{"user:@abc"},
			err:   "missing password",
		},
		{
			name:  "Missing password",
			input: []string{"user@abc"},
			err:   "missing password",
		},
		{
			name:  "Missing host",
			input: []string{"user:pass"},
			err:   "invalid URL",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			_, err := newUserInfoMatcher(tc.input, stdLogger{})
			if err == nil {
				t.Fatal("expected error")
			}
			t.Log(err)
			if !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("expected error containing %s, got %s", tc.err, err)
			}
		})
	}
}

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
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			m, err := newUserInfoMatcher(tc.input, stdLogger{})
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
