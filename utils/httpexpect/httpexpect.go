// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpexpect

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"testing"
)

// NewURLWithBasicAuth parses rawURL and adds user with basicAuth to it.
// If basicAuth is empty, it returns rawURL as is.
func NewURLWithBasicAuth(rawURL, basicAuth string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if basicAuth != "" {
		user, pass, ok := strings.Cut(basicAuth, ":")
		if !ok {
			return nil, fmt.Errorf("invalid basic auth string %q", basicAuth)
		}
		u.User = url.UserPassword(user, pass)
	}

	return u, nil
}

type Client struct {
	t       *testing.T
	rt      http.RoundTripper
	trace   *httptrace.ClientTrace
	baseURL string
}

func NewClient(t *testing.T, baseURL string, rt http.RoundTripper) *Client {
	t.Helper()

	return &Client{
		t:       t,
		rt:      rt,
		baseURL: baseURL,
	}
}

func (c *Client) Trace(enabled bool) {
	if !enabled {
		c.trace = nil
	} else {
		c.trace = newTestClientTrace(c.t)
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.trace != nil {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), c.trace))
	}
	resp, err := c.rt.RoundTrip(req)

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
	c.t.Helper()
	return c.Request("GET", path, opts...)
}

func (c *Client) HEAD(path string, opts ...func(*http.Request)) *Response {
	c.t.Helper()
	return c.Request("HEAD", path, opts...)
}

func (c *Client) Request(method, path string, opts ...func(*http.Request)) *Response {
	c.t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, fmt.Sprintf("%s%s", c.baseURL, path), http.NoBody)
	if err != nil {
		c.t.Fatalf("Failed to create request %s, %s: %v", method, req.URL, err)
	}
	for _, opt := range opts {
		opt(req)
	}
	resp, err := c.do(req)
	if err != nil {
		var tlsErr *tls.CertificateVerificationError
		if errors.As(err, &tlsErr) {
			for i, u := range tlsErr.UnverifiedCertificates {
				c.t.Logf("Unverified certificate[%d]: %s %s %s", i, u.Subject, u.Issuer, u.DNSNames)
			}
		}

		c.t.Fatalf("Failed to execute request %s, %s: %v", method, req.URL, err)
	}

	rr := c.MakeResponse(resp)
	resp.Body.Close()
	return rr
}

func (c *Client) MakeResponse(resp *http.Response) *Response {
	c.t.Helper()
	req := resp.Request
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("Failed to read body from %s, %s: %v", req.Method, req.URL, err)
	}
	return &Response{Response: resp, Body: b, t: c.t}
}

type Response struct {
	*http.Response
	Body []byte
	t    *testing.T
}

func (r *Response) ExpectStatus(status int) *Response {
	if r.StatusCode != status {
		r.t.Fatalf("%s, %s: expected status %d, got %d", r.Request.Method, r.Request.URL, status, r.StatusCode)
	}
	return r
}

func (r *Response) ExpectHeader(key, value string) *Response {
	if v := r.Header.Get(key); v != value {
		r.t.Fatalf("%s, %s: expected header %s to equal '%s', got '%s'", r.Request.Method, r.Request.URL, key, value, v)
	}
	return r
}

func (r *Response) ExpectBodySize(expectedSize int) *Response {
	if bodySize := len(r.Body); bodySize != expectedSize {
		r.t.Fatalf("%s, %s: expected body size %d, got %d", r.Request.Method, r.Request.URL, expectedSize, bodySize)
	}
	return r
}

func (r *Response) ExpectBodyContent(content string) *Response {
	if b := string(r.Body); b != content {
		r.t.Fatalf("%s, %s: expected body to equal '%s', got '%s'", r.Request.Method, r.Request.URL, content, b)
	}
	return r
}
