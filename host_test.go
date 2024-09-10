// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHostPortValidate(t *testing.T) {
	tests := []struct {
		hp  HostPort
		err string
	}{
		{
			hp: HostPort{
				Host: "foo",
				Port: "80",
			},
		},
		{
			hp: HostPort{
				Host: "127.0.0.1",
				Port: "80",
			},
		},
		{
			hp: HostPort{
				Host: "::1",
				Port: "80",
			},
		},
		{
			hp: HostPort{
				Host: "",
				Port: "80",
			},
			err: "missing host",
		},
		{
			hp: HostPort{
				Host: "foo",
				Port: "",
			},
			err: "missing port",
		},
		{
			hp: HostPort{
				Host: "*",
				Port: "80",
			},
			err: "invalid host",
		},
		{
			hp: HostPort{
				Host: "foo",
				Port: "-1",
			},
			err: "invalid port",
		},
		{
			hp: HostPort{
				Host: "foo",
				Port: "1000000",
			},
			err: "invalid port",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.hp.Host+":"+tc.hp.Port, func(t *testing.T) {
			err := tc.hp.Validate()
			if tc.err == "" {
				if err != nil {
					t.Fatalf("expected success, got %q", err)
				}
			} else if !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("expected error to contain %q, got %q", tc.err, err)
			}
		})
	}
}

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

func TestParseHostPortPair(t *testing.T) {
	tests := []struct {
		input string
		hpp   HostPortPair
	}{
		{
			input: "localhost:80:2001:0db8:0000:0000:0000:ff00:0042:8329:443",
			hpp: HostPortPair{
				Src: HostPort{
					Host: "localhost",
					Port: "80",
				},
				Dst: HostPort{
					Host: "2001:0db8:0000:0000:0000:ff00:0042:8329",
					Port: "443",
				},
			},
		},
		{
			input: "2001:0db8:0000:0000:0000:ff00:0042:8329:443:localhost:80",
			hpp: HostPortPair{
				Src: HostPort{
					Host: "2001:0db8:0000:0000:0000:ff00:0042:8329",
					Port: "443",
				},
				Dst: HostPort{
					Host: "localhost",
					Port: "80",
				},
			},
		},
		{
			input: "::1:80:localhost:443",
			hpp: HostPortPair{
				Src: HostPort{
					Host: "::1",
					Port: "80",
				},
				Dst: HostPort{
					Host: "localhost",
					Port: "443",
				},
			},
		},
		{
			input: "[::1]:80:localhost:443",
			hpp: HostPortPair{
				Src: HostPort{
					Host: "::1",
					Port: "80",
				},
				Dst: HostPort{
					Host: "localhost",
					Port: "443",
				},
			},
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.input, func(t *testing.T) {
			hpp, err := ParseHostPortPair(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(hpp, tc.hpp); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
