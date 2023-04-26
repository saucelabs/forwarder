// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package e2e

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
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

type Client struct {
	t       *testing.T
	tr      *http.Transport
	baseURL string
}

func NewClient(t *testing.T, baseURL string, opts ...func(tr *http.Transport)) *Client {
	t.Helper()

	tr := newTransport(t)
	for _, opt := range opts {
		opt(tr)
	}

	return &Client{
		t:       t,
		tr:      tr,
		baseURL: baseURL,
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
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

func (c *Client) GET(path string, opts ...func(*http.Request)) *Response {
	return c.request("GET", path, opts...)
}

func (c *Client) HEAD(path string, opts ...func(*http.Request)) *Response {
	return c.request("HEAD", path, opts...)
}

func (c *Client) request(method, path string, opts ...func(*http.Request)) *Response {
	req, err := http.NewRequestWithContext(context.Background(), method, fmt.Sprintf("%s%s", c.baseURL, path), http.NoBody)
	if err != nil {
		c.t.Fatalf("Failed to create request: %v", err)
	}
	for _, opt := range opts {
		opt(req)
	}
	resp, err := c.do(req)
	if err != nil {
		c.t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("Failed to read response body: %v", err)
	}
	return &Response{Response: resp, body: b, t: c.t}
}

type Response struct {
	*http.Response
	body []byte
	t    *testing.T
}

func (r *Response) ExpectStatus(status int) *Response {
	if r.StatusCode != status {
		r.t.Errorf("Expected status %d, got %d", status, r.StatusCode)
	}
	return r
}

func (r *Response) ExpectBodySize(expectedSize int) *Response {
	if bodySize := len(r.body); bodySize != expectedSize {
		r.t.Errorf("Expected body size %d, got %d", expectedSize, bodySize)
	}
	return r
}

func (r *Response) ExpectBodyContent(content string) *Response {
	if b := string(r.body); b != content {
		r.t.Errorf("Expected body to equal '%s', got: '%s'", content, b)
	}
	return r
}
