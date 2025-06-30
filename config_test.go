// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestParseUserinfo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "normal",
			input: "user:pass",
		},
		{
			name:  "not URL encoded",
			input: "%40:%3A",
		},
		{
			name:  "no user",
			input: ":pass",
			err:   "username cannot be empty",
		},
		{
			name:  "empty",
			input: "",
			err:   "expected username[:password]",
		},
		{
			name:  "two colons",
			input: "user:pass:pass",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			ui, err := ParseUserinfo(tc.input)
			if tc.err == "" {
				if err != nil {
					t.Fatalf("expected success, got %q", err)
				}
				pass, ok := ui.Password()
				if ok {
					pass = ":" + pass
				}
				if ui.Username()+pass != tc.input {
					t.Errorf("expected %q, got %q", tc.input, ui.String())
				}
			} else if !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("expected error to contain %q, got %q", tc.err, err)
			}
		})
	}
}

func TestParseProxyURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "normal",
			input: "192.188.1.100:1080",
		},
		{
			name:  "https",
			input: "https://192.188.1.100:1080",
		},
		{
			name:  "unsupported scheme",
			input: "tcp://192.188.1.100:1080",
			err:   "unsupported scheme",
		},
		{
			name:  "no port",
			input: "192.188.1.100",
			err:   "port is required",
		},
		{
			name:  "port 0",
			input: "192.188.1.100:0",
			err:   "port cannot be 0",
		},
		{
			name:  "hostname",
			input: "saucelabs.com:1080",
		},
		{
			name:  "invalid host name",
			input: "foo-:1080",
			err:   "unable to parse IP",
		},
		{
			name:  "invalid IP",
			input: "1.2.3.400:1080",
			err:   "field has value >255",
		},
		{
			name:  "path",
			input: "192.188.1.100:1080/path",
			err:   "unsupported URL elements",
		},
		{
			name:  "user info",
			input: "http://user%:pass!@1.2.3.4:1080",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseProxyURL(tc.input)
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}

				t.Logf("got error: %s", err)

				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
			}

			if tc.err != "" {
				t.Fatalf("expected error %q, got success", tc.err)
			}
		})
	}
}

func TestParseDNSAddress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "normal",
			input: "1.1.1.1:53",
		},
		{
			name:  "no port",
			input: "1.1.1.1",
		},
		{
			name:  "ipv6",
			input: "[2606:4700:4700::1111]:53",
		},
		{
			name:  "invalid ip",
			input: "300.300.300.300:53",
			err:   "IPv4 field has value >255",
		},
		{
			name:  "invalid port",
			input: "1.1.1.1:abc",
			err:   "invalid syntax",
		},
		{
			name:  "invalid port",
			input: "1.1.1.1:9999999",
			err:   "value out of range",
		},
		{
			name:  "hostname",
			input: "saucelabs.com:53",
			err:   "unexpected character",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			u, err := ParseDNSAddress(tc.input)
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}

				t.Logf("got error: %s", err)

				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
			}

			if tc.err != "" {
				t.Fatalf("expected error %q, got success", tc.err)
			}

			if tc.name == "no port" {
				tc.input += ":53"
			}

			if u.String() != tc.input {
				t.Errorf("expected %q, got %q", tc.input, u.String())
			}
		})
	}
}

func TestParseFilePath(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "temp",
			input: tempDir + "/foo",
		},
		{
			name:  "temp subdirectory",
			input: tempDir + "/a/b/c/foo",
		},

		{
			name:  "insufficient permissions",
			input: "/path/to/file",
			err:   "mkdir",
		},
		{
			name:  "empty",
			input: "",
		},
	}

	for i := range tests {
		tc := &tests[i]
		if runtime.GOOS == "windows" {
			// Windows doesn't support permissions, so we can't test this.
			tc.err = ""
		}

		t.Run(tc.name, func(t *testing.T) {
			f, err := OpenFileParser(os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600, 0o700)(tc.input)
			defer func() {
				if f == nil {
					return
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			}()
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}

				t.Logf("got error: %s", err)

				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
			}

			if tc.err != "" {
				t.Fatalf("expected error %q, got success", tc.err)
			}
		})
	}
}
