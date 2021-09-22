// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package validation

import (
	"testing"
)

func TestSetup(t *testing.T) {
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
			name:    "Should work - scheme http",
			text:    "http://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - scheme https",
			text:    "https://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - scheme socks",
			text:    "socks://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - scheme socks5",
			text:    "socks5://localhost:80",
			wantErr: false,
		},
		{
			name:    "Should work - scheme quic",
			text:    "quic://localhost:80",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Setup()

			if err := v.Var(tt.text, "proxyURL"); (err != nil) != tt.wantErr {
				t.Errorf("Expected %v got %v", tt.wantErr, err)
			}
		})
	}
}
