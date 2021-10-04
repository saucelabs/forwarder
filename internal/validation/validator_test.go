// Copyright 2021 The pacman Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package validation

import (
	"testing"
)

func TestSetup_proxyURIValidator(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "Should work - port low",
			text:    "http://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - port high",
			text:    "http://localhost:65535",
			wantErr: false,
		},
		{
			name:    "Should work - localhost",
			text:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "Should work - IP",
			text:    "http://0.0.0.0:8080",
			wantErr: false,
		},
		{
			name:    "Should work - URL",
			text:    "http://example.com:8080",
			wantErr: false,
		},
		{
			name:    "Should work - port low",
			text:    "https://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - port high",
			text:    "https://localhost:65535",
			wantErr: false,
		},
		{
			name:    "Should work - localhost",
			text:    "https://localhost:8080",
			wantErr: false,
		},
		{
			name:    "Should work - IP",
			text:    "https://0.0.0.0:8080",
			wantErr: false,
		},
		{
			name:    "Should work - URL",
			text:    "https://example.com:8080",
			wantErr: false,
		},
		{
			name:    "Should work - port low",
			text:    "socks://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - port high",
			text:    "socks://localhost:65535",
			wantErr: false,
		},
		{
			name:    "Should work - localhost",
			text:    "socks://localhost:8080",
			wantErr: false,
		},
		{
			name:    "Should work - IP",
			text:    "socks://0.0.0.0:8080",
			wantErr: false,
		},
		{
			name:    "Should work - URL",
			text:    "socks://example.com:8080",
			wantErr: false,
		},

		{
			name:    "Should work - port low",
			text:    "socks5://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - port high",
			text:    "socks5://localhost:65535",
			wantErr: false,
		},
		{
			name:    "Should work - localhost",
			text:    "socks5://localhost:8080",
			wantErr: false,
		},
		{
			name:    "Should work - IP",
			text:    "socks5://0.0.0.0:8080",
			wantErr: false,
		},
		{
			name:    "Should work - URL",
			text:    "socks5://example.com:8080",
			wantErr: false,
		},
		{
			name:    "Should work - port low",
			text:    "quic://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - port high",
			text:    "quic://localhost:65535",
			wantErr: false,
		},
		{
			name:    "Should work - localhost",
			text:    "quic://localhost:8080",
			wantErr: false,
		},
		{
			name:    "Should work - IP",
			text:    "quic://0.0.0.0:8080",
			wantErr: false,
		},
		{
			name:    "Should work - URL",
			text:    "quic://example.com:8080",
			wantErr: false,
		},
		{
			name:    "Should fail - unknown scheme",
			text:    "asd://localhost:80",
			wantErr: true,
		},
		{
			name:    "Should fail - out-of-range low",
			text:    "https://localhost:79",
			wantErr: true,
		},
		{
			name:    "Should fail - out-of-range high",
			text:    "https://localhost:65536",
			wantErr: true,
		},
		{
			name:    "Should fail - empty scheme",
			text:    "localhost:65536",
			wantErr: true,
		},
		{
			name:    "Should fail - empty hostname",
			text:    "http://:65536",
			wantErr: true,
		},
		{
			name:    "Should fail - empty port",
			text:    "http://localhost:",
			wantErr: true,
		},
		{
			name:    "Should fail - invalid URL",
			text:    "::",
			wantErr: true,
		},
		{
			name:    "Should fail - invalid hostname",
			text:    "http://as:65536",
			wantErr: true,
		},
		{
			name:    "Should fail - missing content",
			text:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Get()

			if err := v.Var(tt.text, "proxyURI"); (err != nil) != tt.wantErr {
				t.Errorf("Expected %v got %v", tt.wantErr, err)
			}
		})
	}
}

func TestSetup_pacTextOrURIValidator(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "Should work - http:// - URL",
			text:    "http://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - https:// - URL",
			text:    "https://localhost:65535",
			wantErr: false,
		},
		{
			name:    "Should work - http:// - IP",
			text:    "http://127.0.0.1:80",
			wantErr: false,
		},
		{
			name:    "Should work - https:// - IP",
			text:    "https://127.0.0.1:65535",
			wantErr: false,
		},
		{
			name:    "Should work - file:// - URL",
			text:    "file://localhost:8080",
			wantErr: false,
		},
		{
			name:    "Should work - file:/// - URL",
			text:    "file:///localhost:8080",
			wantErr: false,
		},
		{
			name:    "Should work - file:// - IP",
			text:    "file://127.0.0.1:8080",
			wantErr: false,
		},
		{
			name:    "Should work - file:/// - IP",
			text:    "file:///127.0.0.1:8080",
			wantErr: false,
		},
		{
			name:    "Should work - file:// - Path without extension",
			text:    "file://somedir/somefile:8080",
			wantErr: false,
		},
		{
			name:    "Should work - file:/// - Path without extension",
			text:    "file:///somedir/somefile:8080",
			wantErr: false,
		},
		{
			name:    "Should work - file:// - Path with extension",
			text:    "file://somedir/somefile.pac:8080",
			wantErr: false,
		},
		{
			name:    "Should work - file:/// - Path with extension",
			text:    "file:///somedir/somefile.pac:8080",
			wantErr: false,
		},
		{
			name:    "Should work - Just extension (requires .pac)",
			text:    "somefile.pac",
			wantErr: false,
		},
		{
			name:    "Should work - function FindProxyForURL",
			text:    "function FindProxyForURL(url, host) {}",
			wantErr: false,
		},
		{
			name:    "Should fail - Min length",
			text:    "asd",
			wantErr: true,
		},
		{
			name:    "Should fail - missing any valid keyword",
			text:    "something",
			wantErr: true,
		},
		{
			name:    "Should fail - missing content",
			text:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Setup()

			if err := v.Var(tt.text, "pacTextOrURI"); (err != nil) != tt.wantErr {
				t.Errorf("Expected %v got %v", tt.wantErr, err)
			}
		})
	}
}

func TestSetup_basicAuthCredentialValidator(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{
			name:    "Should work",
			text:    "username:password",
			wantErr: false,
		},
		{
			name:    "Should fail - empty",
			text:    "",
			wantErr: true,
		},
		{
			name:    "Should fail - total less than 7",
			text:    "as",
			wantErr: true,
		},
		{
			name:    "Should fail - missing :",
			text:    "username",
			wantErr: true,
		},
		{
			name:    "Should fail - not 2 components",
			text:    "username:password:something",
			wantErr: true,
		},
		{
			name:    "Should fail - username less than 3",
			text:    ":password",
			wantErr: true,
		},
		{
			name:    "Should fail - password less than 3",
			text:    "username:",
			wantErr: true,
		},
		{
			name:    "Should fail - only :`",
			text:    ":",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Setup()

			if err := v.Var(tt.text, "basicAuth"); (err != nil) != tt.wantErr {
				t.Errorf("Expected %v got %v", tt.wantErr, err)
			}
		})
	}
}
