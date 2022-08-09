// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeURI(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
		err      error
	}{
		{
			name:     "Adds http scheme",
			url:      "example.com",
			expected: "http://example.com",
			err:      nil,
		},
		{
			name:     "Adds http scheme",
			url:      "example.com:8888",
			expected: "http://example.com:8888",
			err:      nil,
		},
		{
			name:     "Adds http scheme",
			url:      "://example.com:8888",
			expected: "http://example.com:8888",
			err:      nil,
		},
		{
			name:     "Adds https scheme",
			url:      "://example.com:443",
			expected: "https://example.com:443",
			err:      nil,
		},
		{
			name:     "Adds https scheme",
			url:      "example.com:443",
			expected: "https://example.com:443",
			err:      nil,
		},
		{
			name:     "Preserves the scheme",
			url:      "https://example.com",
			expected: "https://example.com",
			err:      nil,
		},
	}

	for _, tc := range testCases {
		result, err := NormalizeURI(tc.url)
		if tc.err == nil {
			assert.Equalf(t, tc.expected, result.String(), "%s: Unexpected result: %v", tc.name, result)
			assert.NoErrorf(t, err,
				"%s: Unexpected error: %s", tc.name, err)
		} else {
			assert.Errorf(t, err, "%s: Expected error: %s", tc.name, tc.err)
		}
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
