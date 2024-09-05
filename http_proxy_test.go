// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
	defer p.Close()

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
		s := httptest.NewServer(p.handler())
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
			Handler: p.handler(),
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

func TestNopDialer(t *testing.T) {
	nopDialerErr := errors.New("nop dialer")

	tr := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return nil, nopDialerErr
		},
	}

	p, err := NewHTTPProxy(DefaultHTTPProxyConfig(), nil, nil, tr, stdlog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	req := &http.Request{
		Method: http.MethodGet,
		Header: map[string][]string{},
		URL: &url.URL{
			Scheme: "http",
			Host:   "foobar",
		},
		Host: "foobar",
	}
	_, err = p.proxy.RoundTripper.RoundTrip(req)
	if !errors.Is(err, nopDialerErr) {
		t.Fatalf("expected %v, got %v", nopDialerErr, err)
	}
}

func TestIsLocalhost(t *testing.T) {
	cfg := DefaultHTTPProxyConfig()
	p, err := NewHTTPProxy(cfg, nil, nil, nil, stdlog.Default())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		host      string
		localhost bool
	}{
		{"127.0.0.1", true},
		{"127.10.20.30", true},
		{"localhost", true},
		{"0.0.0.0", true},

		{"notlocalhost", false},
		{"broadcasthost", false},

		{"::1", true},
		{"::", true},

		{"::10", false},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
	}

	for i := range tests {
		tc := tests[i]
		if lh := p.isLocalhost(tc.host); lh != tc.localhost {
			t.Errorf("isLocalhost(%q) = %v; want %v", tc.host, lh, tc.localhost)
		}
	}
}
