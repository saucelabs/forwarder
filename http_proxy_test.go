// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/saucelabs/forwarder/log/stdlog"
	"golang.org/x/net/http2"
)

func TestAbortIf(t *testing.T) {
	// Create proxy with basic auth.
	cfg := DefaultHTTPProxyConfig()
	cfg.BasicAuth = url.UserPassword("user", "pass")

	p, err := NewHTTPProxy(cfg, nil, nil, nil, stdlog.Default())
	if err != nil {
		t.Fatal(err)
	}

	check := func(t *testing.T, rt http.RoundTripper) {
		t.Helper()

		req, err := http.NewRequest(http.MethodGet, "http://foobar", http.NoBody)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusProxyAuthRequired {
			t.Fatalf("expected %d, got %d", http.StatusProxyAuthRequired, resp.StatusCode)
		}
	}

	t.Run("http", func(t *testing.T) {
		s := httptest.NewServer(p.Handler())
		defer s.Close()

		tr := &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		}
		check(t, tr)
	})

	t.Run("http2", func(t *testing.T) {
		c0, c1 := net.Pipe()
		defer c0.Close()
		defer c1.Close()

		var s http2.Server
		go s.ServeConn(c1, &http2.ServeConnOpts{
			Handler: p.Handler(),
		})

		var tr http2.Transport
		tr.AllowHTTP = true
		cc, err := tr.NewClientConn(c0)
		if err != nil {
			t.Fatal(err)
		}
		defer cc.Close()

		check(t, cc)
	})
}

func TestHeaderByPrefixRemoverModifyRequest(t *testing.T) {
	withHeader := func(header http.Header) *http.Request {
		req, err := http.NewRequest(http.MethodGet, "http://example", nil) //nolint:gocritic // This is header test.
		if err != nil {
			t.Fatal(err)
		}
		req.Header = header
		return req
	}

	tests := []struct {
		name     string
		prefix   string
		req      *http.Request
		expected http.Header
	}{
		{
			name:   "smoke",
			prefix: http.CanonicalHeaderKey("RemoveMe"),
			req: withHeader(http.Header{
				http.CanonicalHeaderKey("RemoveMeByPrefix"): nil,
				http.CanonicalHeaderKey("RemoveMeBy"):       nil,
				http.CanonicalHeaderKey("RemoveMe"):         nil,
				http.CanonicalHeaderKey("DontRemoveMe"):     nil,
			}),
			expected: http.Header{
				http.CanonicalHeaderKey("DontRemoveMe"): nil,
			},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			mod := newHeaderRemover(tc.prefix)
			req := tc.req
			err := mod.ModifyRequest(req)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(req.Header, tc.expected); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
