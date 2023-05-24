// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package e2e

import (
	"crypto/tls"
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
	proxy              = serviceScheme("FORWARDER_PROTOCOL") + "://proxy:3128"
	httpbin            = serviceScheme("HTTPBIN_PROTOCOL") + "://httpbin:8080"
	basicAuth          = os.Getenv("FORWARDER_BASIC_AUTH")
	insecureSkipVerify = os.Getenv("INSECURE") != "false"
)

func newTransport(tb testing.TB) *http.Transport {
	tb.Helper()

	tr := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // we know it's a *http.Transport
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec // This is for testing only.
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

	return httpexpect.NewClient(t, baseURL, tr)
}
