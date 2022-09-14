// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeURLScheme(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "Adds http scheme",
			url:      "example.com",
			expected: "http://example.com",
		},
		{
			name:     "Adds http scheme",
			url:      "example.com:8888",
			expected: "http://example.com:8888",
		},
		{
			name:     "Adds http scheme",
			url:      "://example.com:8888",
			expected: "http://example.com:8888",
		},
		{
			name:     "Adds https scheme",
			url:      "://example.com:443",
			expected: "https://example.com:443",
		},
		{
			name:     "Adds https scheme",
			url:      "example.com:443",
			expected: "https://example.com:443",
		},
		{
			name:     "Preserves the scheme",
			url:      "https://example.com",
			expected: "https://example.com",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(normalizeURLScheme(tc.url))
			if err != nil {
				t.Fatal(err)
			}
			if u.String() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, u)
			}
		})
	}
}

func TestIsLocalHost(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Is localhost",
			url:      "http://localhost:8080",
			expected: true,
		},
		{
			name:     "Is localhost IPv6",
			url:      "http://[::1]:8080",
			expected: true,
		},
		{
			name:     "Is localhost IPv6 no port",
			url:      "http://[::1]",
			expected: true,
		},
		{
			name:     "Is localhost IPv4",
			url:      "http://127.0.0.2:8080",
			expected: true,
		},
		{
			name:     "Not localhost",
			url:      "example.com:8888",
			expected: false,
		},
	}

	for _, tc := range testCases {
		req, err := http.NewRequest("GET", tc.url, nil)
		assert.NoErrorf(t, err, "%s: Unexpected error creating request for :%s", tc.name, tc.url)
		isLocalhost := IsLocalHost(req)
		assert.Equalf(t, tc.expected, isLocalhost, "%s: Unexpected result: %v", tc.name, isLocalhost)
	}
}
