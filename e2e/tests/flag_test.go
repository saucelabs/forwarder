// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/e2e/forwarder"
)

func TestFlagProxyLocalhost(t *testing.T) {
	hosts := []string{
		"localhost",
		"127.0.0.1",
	}

	t.Run("allow", func(t *testing.T) {
		for _, h := range hosts {
			newClient(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").ExpectStatus(http.StatusOK)
		}
	})
	t.Run("deny", func(t *testing.T) {
		for _, h := range hosts {
			newClient(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").ExpectStatus(http.StatusForbidden)
		}
	})
}

func TestFlagHeader(t *testing.T) {
	c := newClient(t, httpbin)

	c.GET("/headers/").ExpectStatus(http.StatusOK).ExpectHeader("test-add", "test-value")

	c.GET("/headers/", setHeader("test-empty", "not-empty")).ExpectStatus(http.StatusOK).ExpectHeader("test-empty", "")

	c.GET("/headers/", setHeader("test-rm", "value-1")).ExpectStatus(http.StatusOK).ExpectHeader("test-rm", "")

	c.GET("/headers/", setHeader("rm-prefix", "value-2")).ExpectStatus(http.StatusOK).ExpectHeader("rm-prefix", "")
}

func TestFlagResponseHeader(t *testing.T) {
	c := newClient(t, httpbin)

	c.GET("/status/200").ExpectStatus(http.StatusOK).ExpectHeader("test-resp-add", "test-resp-value")

	c.GET("/headers/", setHeader("test-resp-empty", "not-empty")).ExpectStatus(http.StatusOK).ExpectHeader("test-resp-empty", "")

	c.GET("/headers/", setHeader("test-resp-rm", "value-3")).ExpectStatus(http.StatusOK).ExpectHeader("test-resp-rm", "")

	c.GET("/headers/", setHeader("resp-rm-prefix", "value-4")).ExpectStatus(http.StatusOK).ExpectHeader("resp-rm-prefix", "")
}

func setHeader(key, value string) func(r *http.Request) {
	return func(r *http.Request) {
		r.Header.Set(key, value)
	}
}

var httpbinDNS = serviceScheme("HTTPBIN_PROTOCOL") + "://httpbin.local:8080"

func TestFlagDNSServer(t *testing.T) {
	t.Run("default httpbin address", func(t *testing.T) {
		newClient(t, httpbin).GET("/status/200").ExpectStatus(http.StatusBadGateway)
	})

	t.Run("custom httpbin address", func(t *testing.T) {
		newClient(t, httpbinDNS).GET("/status/200").ExpectStatus(http.StatusOK)
	})
}

func TestFlagInsecure(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		newClient(t, httpbin).GET("/status/200").ExpectStatus(http.StatusOK)
	})
	t.Run("false", func(t *testing.T) {
		for _, scheme := range []string{"http", "https"} {
			newClient(t, scheme+"://httpbin:8080").GET("/status/200").ExpectStatus(http.StatusBadGateway)
		}
	})
}

func TestFlagMITM(t *testing.T) {
	newClient(t, httpbin).GET("/status/200").ExpectStatus(http.StatusOK).
		ExpectHeader("test-resp-add", "test-resp-value")
}

func TestFlagMITMGenCA(t *testing.T) {
	r := newClient(t, proxyAPI, func(tr *http.Transport) {
		tr.Proxy = nil
	}).GET("/cacert").ExpectStatus(http.StatusOK)

	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatal(err)
	}
	if ok := pool.AppendCertsFromPEM(r.Body); !ok {
		t.Fatal("failed to append cert from response")
	}

	newClient(t, httpbin, func(tr *http.Transport) {
		tr.TLSClientConfig.RootCAs = pool
	}).GET("/status/200").
		ExpectStatus(http.StatusOK).
		ExpectHeader("test-resp-add", "test-resp-value")
}

func TestFlagMITMDomains(t *testing.T) {
	t.Run("include(google)", func(t *testing.T) {
		newClient(t, "https://www.google.com").GET("/").
			ExpectStatus(http.StatusOK).
			ExpectHeader("test-resp-add", "test-resp-value")
	})

	t.Run("exclude implicitly(httpbin)", func(t *testing.T) {
		newClient(t, httpbin).GET("/status/200").
			ExpectStatus(http.StatusOK).
			ExpectHeader("test-resp-add", "")
	})

	t.Run("exclude(amazon)", func(t *testing.T) {
		newClient(t, "https://www.amazon.com").GET("/").
			ExpectStatus(http.StatusOK).
			ExpectHeader("test-resp-add", "")
	})
}

func TestFlagDenyDomains(t *testing.T) {
	t.Run("include(google)", func(t *testing.T) {
		newClient(t, "https://www.google.com").GET("/").
			ExpectStatus(http.StatusForbidden)
	})

	t.Run("exclude(httpbin)", func(t *testing.T) {
		newClient(t, httpbin).GET("/status/200").
			ExpectStatus(http.StatusOK)
	})
}

func TestFlagDirectDomains(t *testing.T) {
	viaHeader := newClient(t, httpbin).GET("/headers/").Header["Via"]
	var success bool
	for _, via := range viaHeader {
		if strings.Contains(via, forwarder.UpstreamProxyServiceName) {
			success = true
		}
	}

	if success {
		t.Fatalf("%s via header found", forwarder.UpstreamProxyServiceName)
	}
}

func TestFlagReadLimit(t *testing.T) {
	const (
		// It streams 2 * 5MiB, the read limit is 1MiB/s, but minimum burst is 4MiB, so it should take approximately 6 seconds.
		connections  = 2
		size         = 5 * 1024 * 1024 // 5MiB
		expectedTime = 6 * time.Second
		epsilon      = expectedTime * 5 / 100 // 5%
	)

	testRateLimitHelper(t, connections, expectedTime, epsilon, func() {
		c := newClient(t, httpbin)
		c.GET(fmt.Sprintf("/stream-bytes/%d", size)).ExpectStatus(http.StatusOK).ExpectBodySize(size)
	})
}

func TestFlagWriteLimit(t *testing.T) {
	const (
		// It streams 2 * 5MiB, the write limit is 1MiB/s, but minimum burst is 4MiB, so it should take approximately 6 seconds.
		connections  = 2
		size         = 5 * 1024 * 1024 // 5MiB
		expectedTime = 6 * time.Second
		epsilon      = expectedTime * 5 / 100 // 5%
	)

	testRateLimitHelper(t, connections, expectedTime, epsilon, func() {
		c := newClient(t, httpbin)
		c.GET("/count-bytes/", func(req *http.Request) {
			req.Header.Set("Content-Type", "application/octet-stream")
			req.Body = &bytesReaderCloser{bytes.NewReader(make([]byte, size))}
		}).ExpectStatus(http.StatusOK).ExpectHeader("body-size", fmt.Sprintf("%d", size))
	})
}

type bytesReaderCloser struct {
	*bytes.Reader
}

func (b *bytesReaderCloser) Close() error {
	return nil
}

func testRateLimitHelper(t *testing.T, workers int, expectedTime, epsilon time.Duration, cb func()) {
	t.Helper()

	var wg sync.WaitGroup

	ts := time.Now()
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb()
		}()
	}
	wg.Wait()
	if elapsed := time.Since(ts); elapsed < expectedTime-epsilon || elapsed > expectedTime+epsilon {
		t.Fatalf("Expected request to take approximately %s, took %s", expectedTime, elapsed)
	}
}
