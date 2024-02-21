// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestProxyResolverChromium(t *testing.T) { //nolint:maintidx // long table
	defaultQueryURL, err := url.ParseRequestURI("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		fileName  string
		configure func(t *testing.T, cfg *ProxyResolverConfig)
		queryURL  string
		want      []Proxy
		err       string
		evalErr   string
	}{
		{
			fileName: "ambiguous_entry_point.js",
			err:      "ambiguous entry point",
		},
		{
			fileName: "b_132073833.js",
			want:     []Proxy{{Mode: DIRECT}},
		},
		{
			fileName: "b_139806216.js",
			want:     []Proxy{{Mode: DIRECT}},
		},
		{
			fileName: "b_147664838.js",
			want:     []Proxy{{Mode: DIRECT}},
		},
		{
			fileName: "binding_from_global.js",
			configure: func(t *testing.T, cfg *ProxyResolverConfig) {
				cfg.testingMyIPAddress = []net.IP{net.ParseIP("1.2.3.4")}
			},
			want: []Proxy{{Mode: PROXY, Host: "1.2.3.4", Port: "80"}},
		},
		{
			fileName: "bindings.js",
			configure: func(t *testing.T, cfg *ProxyResolverConfig) {
				cfg.testingLookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
					return []net.IP{net.ParseIP("127.0.0.1")}, nil
				}
			},
			want: []Proxy{{Mode: DIRECT}},
		},
		// The "change_element_kind.js" test DOES NOT WORK.
		// Goja does not implement the deprecated __defineGetter__ and __defineSetter__ methods.
		// The code fails with "TypeError: Object has no member '__defineGetter__'
		// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object/__defineGetter__
		{
			fileName: "direct.js",
			want:     []Proxy{{Mode: DIRECT}},
		},
		{
			fileName: "dns_fail.js",
			configure: func(t *testing.T, cfg *ProxyResolverConfig) {
				cfg.testingLookupIP = func(ctx context.Context, network, host string) ([]net.IP, error) {
					return nil, errors.New("test")
				}
				cfg.testingMyIPAddress = []net.IP{}
				cfg.testingMyIPAddressEx = []net.IP{}
			},
			want: []Proxy{{Mode: PROXY, Host: "success", Port: "80"}},
		},
		{
			fileName: "ends_with_comment.js",
			want:     []Proxy{{Mode: PROXY, Host: "success", Port: "80"}},
		},
		{
			fileName: "ends_with_statement_no_semicolon.js",
			want:     []Proxy{{Mode: PROXY, Host: "success", Port: "3"}},
		},
		// The "international_domain_names.js" test DOES NOT WORK.
		// It relays on particular side effects which we do not have.
		// However, domains are properly resolved thanks to Golang's SDK.
		{
			fileName: "missing_close_brace.js",
			err:      "Unexpected end of input",
		},
		{
			fileName: "no_entrypoint.js",
			err:      "missing required function FindProxyForURL or FindProxyForURLEx",
		},
		{
			fileName: "pac_library_unittest.js",
			want:     []Proxy{{Mode: PROXY, Host: "success", Port: "80"}},
		},
		{
			fileName: "passthrough.js",
			queryURL: "http://query.com/path",
			want:     []Proxy{{Mode: PROXY, Host: "http.query.com.path.query.com", Port: "80"}},
		},
		{
			fileName: "passthrough.js",
			queryURL: "ftp://query.com:90/path",
			want:     []Proxy{{Mode: PROXY, Host: "ftp.query.com.90.path.query.com", Port: "80"}},
		},
		{
			fileName: "return_empty_string.js",
		},
		{
			fileName: "return_function.js",
			evalErr:  "unexpected return type",
		},
		{
			fileName: "return_integer.js",
			evalErr:  "unexpected return type",
		},
		{
			fileName: "return_null.js",
			evalErr:  "unexpected return type",
		},
		{
			fileName: "return_object.js",
			evalErr:  "unexpected return type",
		},
		{
			fileName: "return_undefined.js",
			evalErr:  "unexpected return type",
		},
		{
			fileName: "return_unicode.js",
			evalErr:  "non-ASCII characters in the return value",
		},
		// SKIP "side_effects.js" test.
		// It requires calling FindProxyForURL 3 times and provides little value compatibility-wise.
		{
			fileName: "simple.js",
			configure: func(t *testing.T, cfg *ProxyResolverConfig) {
				cfg.testingMyIPAddress = []net.IP{net.ParseIP("172.16.3.4")}
			},
			want: []Proxy{{Mode: PROXY, Host: "a", Port: "80"}},
		},
		{
			fileName: "simple.js",
			queryURL: "http://10.2.3.4/path",
			want:     []Proxy{{Mode: PROXY, Host: "b", Port: "80"}},
		},
		{
			fileName: "simple.js",
			queryURL: "http://x.foo.bar.baz.com/path",
			want:     []Proxy{{Mode: PROXY, Host: "c", Port: "100"}},
		},
		{
			fileName: "string_functions.js",
			want:     []Proxy{{Mode: DIRECT}},
		},
		{
			fileName: "unhandled_exception.js",
			evalErr:  "undefined_variable is not defined",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.fileName, func(t *testing.T) {
			b, err := os.ReadFile("testdata/chromium-libpac/" + tc.fileName)
			if err != nil {
				t.Fatal(err)
			}

			var alerts bytes.Buffer
			defer func() {
				if t.Failed() && alerts.Len() > 0 {
					t.Log(alerts.String())
				}
			}()

			cfg := &ProxyResolverConfig{
				Script:    string(b),
				AlertSink: &alerts,
			}
			if tc.configure != nil {
				tc.configure(t, cfg)
			}

			pr, err := NewProxyResolver(cfg, nil)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err.Error())
				}

				t.Log("NewProxyResolver error:", err)
				return
			} else if err != nil {
				t.Fatal(err)
			}

			q := defaultQueryURL
			if tc.queryURL != "" {
				q, err = url.ParseRequestURI(tc.queryURL)
				if err != nil {
					t.Fatal(err)
				}
			}

			p, err := pr.FindProxyForURL(q, "")
			if tc.evalErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tc.evalErr) {
					t.Fatalf("expected error to contain %q, got %q", tc.evalErr, err.Error())
				}

				t.Log("FindProxyForURL error:", err)
				return
			} else if err != nil {
				t.Fatal(err)
			}

			t.Log("FindProxyForURL:", p)

			got, err := Proxies(p).All()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("unexpected proxy list (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProxyResolverLibpac(t *testing.T) {
	tests := []int{
		1,
		2,
		3,
		// SKIP 4 - requires FindProxyForURL to be called 100 times with different URL path
		5,
		6,
		// SKIP 7 - fails with got "PROXY null:8080", want "PROXY :8080"
	}
	for _, i := range tests {
		testFile := fmt.Sprintf("test%d.sh", i)
		t.Run(testFile, func(t *testing.T) {
			pacFile, calls := readLibpacTestShellScript(t, "testdata/libpac/"+testFile)
			if pacFile == "" {
				t.Fatal("no PACFILE found")
			}
			if len(calls) == 0 {
				t.Fatal("no calls found")
			}

			b, err := os.ReadFile("testdata/libpac/" + pacFile)
			if err != nil {
				t.Fatal(err)
			}

			var alerts bytes.Buffer
			defer func() {
				if t.Failed() && alerts.Len() > 0 {
					t.Log(alerts.String())
				}
			}()

			cfg := &ProxyResolverConfig{
				Script:    string(b),
				AlertSink: &alerts,
			}

			pr, err := NewProxyResolver(cfg, nil)
			if err != nil {
				t.Fatal(err)
			}

			for _, c := range calls {
				if c.hostname != "" {
					t.Logf("using custom host name %q for %q", c.hostname, c.url)
				}

				p, err := pr.FindProxyForURL(c.url, c.hostname)
				switch {
				case strings.HasPrefix(c.msg, "Found proxy "):
					if err != nil {
						t.Fatalf("FindProxyForURL(%q) error: %v", c.url, err)
					}

					want := strings.TrimPrefix(c.msg, "Found proxy ")
					if p != want {
						t.Errorf("FindProxyForURL(%q) = %q, want %q", c.url, p, want)
					}
				case strings.HasPrefix(c.msg, "Javascript call failed"):
					if err == nil {
						t.Fatalf("expected error %q", c.msg)
					}
					t.Log("FindProxyForURL error:", err)
				default:
					t.Errorf("unknown call message: %q", c.msg)
				}
			}
		})
	}
}

type libpacTestCall struct {
	url      *url.URL
	hostname string
	msg      string
}

func readLibpacTestShellScript(t *testing.T, path string) (pacFile string, calls []libpacTestCall) {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	pacFileRegex := regexp.MustCompile(`^PACFILE="(\d+.js)"`)

	const (
		progPos = iota + 1
		urlPos
		hostnamePos
		msgPos
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := pacFileRegex.FindStringSubmatch(line); m != nil {
			pacFile = m[1]
			continue
		}

		if strings.HasPrefix(line, "test_") {
			parts := strings.SplitN(line, " ", msgPos+1)
			u, err := url.ParseRequestURI(parts[urlPos])
			if err != nil {
				t.Fatal(err)
			}

			c := libpacTestCall{
				url: u,
				msg: strings.Trim(parts[msgPos], `"'`),
			}
			if h := parts[hostnamePos]; u.Hostname() != h {
				c.hostname = h
			}
			calls = append(calls, c)
		}
	}

	return //nolint:nakedret // pacFile and calls are named return values
}
