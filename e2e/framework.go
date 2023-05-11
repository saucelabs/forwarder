// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package e2e

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	insecureSkipVerify = os.Getenv("INSECURE") != "false"
)

func newProxyURL(tb testing.TB) *url.URL {
	tb.Helper()

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		tb.Fatal(err)
	}

	if ba := os.Getenv("FORWARDER_BASIC_AUTH"); ba != "" {
		u, p, _ := strings.Cut(ba, ":")
		proxyURL.User = url.UserPassword(u, p)
		tb.Log("using basic auth for proxy", proxyURL)
	}

	return proxyURL
}

func newTransport(tb testing.TB) *http.Transport {
	tb.Helper()

	proxyURL := newProxyURL(tb)
	tr := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // we know it's a *http.Transport
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec // This is for testing only.
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
