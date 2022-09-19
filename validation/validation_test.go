// Copyright 2021 The Forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestIsBasicAuth(t *testing.T) {
	tests := []struct {
		name string
		text string
		err  bool
	}{
		{
			name: "Should work",
			text: "username:password",
			err:  false,
		},
		{
			name: "Should fail - empty",
			text: "",
			err:  true,
		},
		{
			name: "Should fail - total less than 7",
			text: "as",
			err:  true,
		},
		{
			name: "Should fail - missing :",
			text: "username",
			err:  true,
		},
		{
			name: "Should work - password with :",
			text: "username:password:something",
			err:  false,
		},
		{
			name: "Should fail - username less than 3",
			text: ":password",
			err:  true,
		},
		{
			name: "Should fail - password less than 3",
			text: "username:",
			err:  true,
		},
		{
			name: "Should fail - only :`",
			text: ":",
			err:  true,
		},
	}

	v := Validator()

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if err := v.Var(tc.text, "basicAuth"); (err != nil) != tc.err {
				t.Errorf("IsBasicAuth() error = %v, expected %v", err, tc.err)
			}
		})
	}
}

func TestIsDNSURI(t *testing.T) {
	tests := []struct {
		name string
		text string
		err  bool
	}{
		{
			name: "Should work - port low",
			text: "udp://localhost:53",
			err:  false,
		},
		{
			name: "Should work - port high",
			text: "udp://localhost:65535",
			err:  false,
		},
		{
			name: "Should work - localhost",
			text: "udp://localhost:8080",
			err:  false,
		},
		{
			name: "Should work - IP",
			text: "udp://0.0.0.0:8080",
			err:  false,
		},
		{
			name: "Should work - URL",
			text: "udp://example.com:8080",
			err:  false,
		},
		{
			name: "Should work - port low",
			text: "tcp://localhost:80",
			err:  false,
		},
		{
			name: "Should work - port high",
			text: "tcp://localhost:65535",
			err:  false,
		},
		{
			name: "Should work - localhost",
			text: "tcp://localhost:8080",
			err:  false,
		},
		{
			name: "Should work - IP",
			text: "tcp://0.0.0.0:8080",
			err:  false,
		},
		{
			name: "Should work - URL",
			text: "tcp://example.com:8080",
			err:  false,
		},
		{
			name: "Should fail - unknown scheme",
			text: "asd://localhost:80",
			err:  true,
		},
		{
			name: "Should fail - out-of-range low",
			text: "tcp://localhost:0",
			err:  true,
		},
		{
			name: "Should fail - out-of-range high",
			text: "tcp://localhost:65536",
			err:  true,
		},
		{
			name: "Should fail - empty scheme",
			text: "localhost:65536",
			err:  true,
		},
		{
			name: "Should fail - empty hostname",
			text: "udp://:65536",
			err:  true,
		},
		{
			name: "Should fail - empty port",
			text: "udp://localhost:",
			err:  true,
		},
		{
			name: "Should fail - invalid URL",
			text: "::",
			err:  true,
		},
		{
			name: "Should fail - invalid hostname",
			text: "udp://as:65536",
			err:  true,
		},
		{
			name: "Should fail - missing content",
			text: "",
			err:  true,
		},
	}

	v := validator.New()
	RegisterAll(v)

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if err := v.Var(tc.text, "dnsURI"); (err != nil) != tc.err {
				t.Errorf("IsDNSURI() error = %v, expected %v", err, tc.err)
			}
		})
	}
}

func TestIsPacURIOrText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "Should work - http:// - URL",
			text:     "http://localhost:80",
			expected: false,
		},
		{
			name:     "Should work - https:// - URL",
			text:     "https://localhost:65535",
			expected: false,
		},
		{
			name:     "Should work - http:// - IP",
			text:     "http://127.0.0.1:80",
			expected: false,
		},
		{
			name:     "Should work - https:// - IP",
			text:     "https://127.0.0.1:65535",
			expected: false,
		},
		{
			name:     "Should work - file:// - URL",
			text:     "file://localhost:8080",
			expected: false,
		},
		{
			name:     "Should work - file:/// - URL",
			text:     "file:///localhost:8080",
			expected: false,
		},
		{
			name:     "Should work - file:// - IP",
			text:     "file://127.0.0.1:8080",
			expected: false,
		},
		{
			name:     "Should work - file:/// - IP",
			text:     "file:///127.0.0.1:8080",
			expected: false,
		},
		{
			name:     "Should work - file:// - Path without extension",
			text:     "file://somedir/somefile:8080",
			expected: false,
		},
		{
			name:     "Should work - file:/// - Path without extension",
			text:     "file:///somedir/somefile:8080",
			expected: false,
		},
		{
			name:     "Should work - file:// - Path with extension",
			text:     "file://somedir/somefile.pac:8080",
			expected: false,
		},
		{
			name:     "Should work - file:/// - Path with extension",
			text:     "file:///somedir/somefile.pac:8080",
			expected: false,
		},
		{
			name:     "Should work - function FindProxyForURL",
			text:     "function FindProxyForURL(url, host) {}",
			expected: false,
		},
		{
			name:     "Should fail - Min length",
			text:     "asd",
			expected: true,
		},
		{
			name:     "Should fail - missing any valid keyword",
			text:     "something",
			expected: true,
		},
		{
			name:     "Should fail - missing content",
			text:     "",
			expected: true,
		},
	}

	v := Validator()

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if err := v.Var(tc.text, "pacURIOrText"); (err != nil) != tc.expected {
				t.Errorf("IsPacURIOrText() error = %v, expected %v", err, tc.expected)
			}
		})
	}
}

func TestIsProxyURI(t *testing.T) {
	tests := []struct {
		name string
		text string
		err  bool
	}{
		{
			name: "Should work - port low",
			text: "http://localhost:80",
			err:  false,
		},
		{
			name: "Should work - port high",
			text: "http://localhost:65535",
			err:  false,
		},
		{
			name: "Should work - localhost",
			text: "http://localhost:8080",
			err:  false,
		},
		{
			name: "Should work - IP",
			text: "http://0.0.0.0:8080",
			err:  false,
		},
		{
			name: "Should work - URL",
			text: "http://example.com:8080",
			err:  false,
		},
		{
			name: "Should work - port low",
			text: "https://localhost:80",
			err:  false,
		},
		{
			name: "Should work - port high",
			text: "https://localhost:65535",
			err:  false,
		},
		{
			name: "Should work - localhost",
			text: "https://localhost:8080",
			err:  false,
		},
		{
			name: "Should work - IP",
			text: "https://0.0.0.0:8080",
			err:  false,
		},
		{
			name: "Should work - URL",
			text: "https://example.com:8080",
			err:  false,
		},
		{
			name: "Should work - port low",
			text: "socks://localhost:80",
			err:  false,
		},
		{
			name: "Should work - port high",
			text: "socks://localhost:65535",
			err:  false,
		},
		{
			name: "Should work - localhost",
			text: "socks://localhost:8080",
			err:  false,
		},
		{
			name: "Should work - IP",
			text: "socks://0.0.0.0:8080",
			err:  false,
		},
		{
			name: "Should work - URL",
			text: "socks://example.com:8080",
			err:  false,
		},

		{
			name: "Should work - port low",
			text: "socks5://localhost:80",
			err:  false,
		},
		{
			name: "Should work - port high",
			text: "socks5://localhost:65535",
			err:  false,
		},
		{
			name: "Should work - localhost",
			text: "socks5://localhost:8080",
			err:  false,
		},
		{
			name: "Should work - IP",
			text: "socks5://0.0.0.0:8080",
			err:  false,
		},
		{
			name: "Should work - URL",
			text: "socks5://example.com:8080",
			err:  false,
		},
		{
			name: "Should work - port low",
			text: "quic://localhost:80",
			err:  false,
		},
		{
			name: "Should work - port high",
			text: "quic://localhost:65535",
			err:  false,
		},
		{
			name: "Should work - localhost",
			text: "quic://localhost:8080",
			err:  false,
		},
		{
			name: "Should work - IP",
			text: "quic://0.0.0.0:8080",
			err:  false,
		},
		{
			name: "Should work - URL",
			text: "quic://example.com:8080",
			err:  false,
		},
		{
			name: "Should fail - unknown scheme",
			text: "asd://localhost:80",
			err:  true,
		},
		{
			name: "Should fail - out-of-range low",
			text: "https://localhost:0",
			err:  true,
		},
		{
			name: "Should fail - out-of-range high",
			text: "https://localhost:65536",
			err:  true,
		},
		{
			name: "Should fail - empty scheme",
			text: "localhost:65536",
			err:  true,
		},
		{
			name: "Should fail - empty hostname",
			text: "http://:65536",
			err:  true,
		},
		{
			name: "Should fail - empty port",
			text: "http://localhost:",
			err:  true,
		},
		{
			name: "Should fail - invalid URL",
			text: "::",
			err:  true,
		},
		{
			name: "Should fail - invalid hostname",
			text: "http://as:65536",
			err:  true,
		},
		{
			name: "Should fail - missing content",
			text: "",
			err:  true,
		},
	}

	v := Validator()

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if err := v.Var(tc.text, "proxyURI"); (err != nil) != tc.err {
				t.Errorf("IsProxyURI() error = %v, wantErr %v", err, tc.err)
			}
		})
	}
}
