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

	"github.com/gavv/httpexpect/v2"
	"github.com/gorilla/websocket"
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

func newTransport(tb testing.TB) *http.Transport {
	tb.Helper()

	if proxy == "" {
		tb.Fatal("proxy URL not set")
	}

	tr := http.DefaultTransport.(*http.Transport).Clone() //nolint:forcetypeassert // we know it's a *http.Transport
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecureSkipVerify, //nolint:gosec // This is for testing only.
	}

	if proxy == "" {
		tb.Log("proxy not set, running without proxy")
	} else {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			tb.Fatal(err)
		}
		if ba := os.Getenv("FORWARDER_BASIC_AUTH"); ba != "" {
			u, p, _ := strings.Cut(ba, ":")
			proxyURL.User = url.UserPassword(u, p)
			tb.Log("using basic auth for proxy", proxyURL)
		}
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	return tr
}

type client struct {
	tr *http.Transport
}

func (c client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.tr.RoundTrip(req)

	// There is a difference between sending HTTP and HTTPS requests.
	// For HTTPS client issues a CONNECT request to the proxy and then sends the original request.
	// In case the proxy responds with status code 4XX or 5XX to the CONNECT request, the client interprets it as URL error.
	//
	// This is to cover this case.
	if req.URL.Scheme == "https" && err != nil {
		for i := 400; i < 600; i++ {
			if err.Error() == http.StatusText(i) {
				return &http.Response{
					StatusCode: i,
					Status:     http.StatusText(i),
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header:     http.Header{},
					Body:       http.NoBody,
					Request:    req,
				}, nil
			}
		}
	}

	return resp, err
}

func Expect(t *testing.T, baseURL string, opts ...func(*httpexpect.Config)) *httpexpect.Expect {
	t.Helper()
	tr := newTransport(t)
	cfg := httpexpect.Config{
		BaseURL:  baseURL,
		Client:   client{tr: tr},
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
		WebsocketDialer: &websocket.Dialer{
			Proxy:           tr.Proxy,
			TLSClientConfig: tr.TLSClientConfig,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return httpexpect.WithConfig(cfg)
}

func ProxyNoAuth(config *httpexpect.Config) {
	tr := config.Client.(client).tr //nolint:forcetypeassert // we know it's a client
	p := tr.Proxy
	tr.Proxy = func(req *http.Request) (u *url.URL, err error) {
		u, err = p(req)
		if u != nil {
			u.User = nil
		}
		return
	}
}
