// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package tests

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/saucelabs/forwarder/utils/httpexpect"
)

func serviceScheme(envVar string) string {
	s := os.Getenv(envVar)
	if s == "h2" {
		return "https"
	}
	if s == "" {
		return "http"
	}
	return s
}

var (
	proxy     = serviceScheme("FORWARDER_PROTOCOL") + "://proxy:3128"
	httpbin   = serviceScheme("HTTPBIN_PROTOCOL") + "://httpbin:8080"
	basicAuth = os.Getenv("FORWARDER_BASIC_AUTH")
)

const caCertFile = "/etc/forwarder/certs/ca.crt"

func defaultTLSConfig() (*tls.Config, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}
	if ok := pool.AppendCertsFromPEM(b); !ok {
		return nil, fmt.Errorf("failed to append cert from %s", caCertFile)
	}

	return &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS13,
	}, nil
}

func newTransport(tb testing.TB) *http.Transport {
	tb.Helper()

	tr := http.DefaultTransport.(*http.Transport).Clone()
	if tlsCfg, err := defaultTLSConfig(); err != nil {
		tb.Fatal(err)
	} else {
		tr.TLSClientConfig = tlsCfg
	}

	proxyURL, err := httpexpect.NewURLWithBasicAuth(proxy, basicAuth)
	if err != nil {
		tb.Fatal(err)
	}
	tr.Proxy = http.ProxyURL(proxyURL)

	return tr
}

func newClient(t *testing.T, baseURL string, opts ...func(tr *http.Transport)) *httpexpect.Client {
	t.Helper()

	tr := newTransport(t)
	for _, opt := range opts {
		opt(tr)
	}

	c := httpexpect.NewClient(t, baseURL, tr)
	// c.Trace(true)
	return c
}
