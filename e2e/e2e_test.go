// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package e2e

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/mmatczuk/anyflag"
	"golang.org/x/sync/errgroup"
)

func TestMain(m *testing.M) {
	flag.Var(anyflag.NewValue[*TestConfig](testConfig, &testConfig, parseTestConfig), "config-file", "config of the test, proxy/httpbin address, etc...")
	if !flag.Parsed() {
		flag.Parse()
	}

	var eg errgroup.Group
	eg.Go(func() error {
		return waitForServerReady(testConfig.ProxyAPI)
	})
	eg.Go(func() error {
		return waitForServerReady(testConfig.HTTPBinAPI)
	})
	if testConfig.UpstreamAPI != "" {
		eg.Go(func() error {
			return waitForServerReady(testConfig.UpstreamAPI)
		})
	}
	if err := eg.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestStatusCodes(t *testing.T) {
	t.Parallel()
	// List of all valid status codes plus some non-standard ones.
	// See https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
	validStatusCodes := []int{
		// FIXME: proxy wrongly supports 1xx, see #113
		// 100, 101, 102, 103, 122,
		200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
		300, 301, 302, 303, 304, 305, 306, 307, 308,
		400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 421, 422, 423, 424, 425, 426, 428, 429, 431, 451,
		500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511, 599,
	}

	e := Expect(t, testConfig.HTTPBin)
	for i := range validStatusCodes {
		code := validStatusCodes[i]
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			t.Parallel()
			e.GET(fmt.Sprintf("/status/%d", code)).Expect().Status(code)
		})
	}
}

func TestAuth(t *testing.T) {
	e := Expect(t, testConfig.HTTPBin)
	t.Run("ok", func(t *testing.T) {
		e.GET("/basic-auth/user/passwd").WithTransformer(func(r *http.Request) {
			r.SetBasicAuth("user", "passwd")
		}).Expect().Status(http.StatusOK)
	})
	t.Run("nok", func(t *testing.T) {
		e.GET("/basic-auth/user/passwd").Expect().Status(http.StatusUnauthorized)
	})
}

func TestProxyAuth(t *testing.T) {
	if testConfig.ProxyBasicAuth == "" {
		t.Skip("proxy basic auth not set")
	}
	e := Expect(t, testConfig.HTTPBin, ProxyNoAuth)
	e.GET("/status/200").Expect().Status(http.StatusProxyAuthRequired)
}

func TestStreamBytes(t *testing.T) {
	t.Parallel()
	var sizes []int
	with := func(size, n int) {
		for i := 0; i < n; i++ {
			sizes = append(sizes, size)
		}
	}

	const base = 100
	with(5, 10*base)
	with(100_000, base)
	with(1_000_000, base/10)

	var (
		workers = runtime.NumCPU()
		wg      sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := Expect(t, testConfig.HTTPBin, func(config *httpexpect.Config) {
				config.Printers = []httpexpect.Printer{}
			})
			for _, p := range rand.Perm(len(sizes)) {
				size := sizes[p]
				e.GET(fmt.Sprintf("/stream-bytes/%d", size)).Expect().Status(http.StatusOK).Body().Length().IsEqual(size)
			}
		}()
	}

	wg.Wait()
}

func TestServerSentEvents(t *testing.T) {
	t.Parallel()

	tr := newTransport(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testConfig.HTTPBin+"/events/100", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	var (
		buf [1024]byte
		i   int
	)
	for {
		n, err := resp.Body.Read(buf[:])
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			t.Fatal(err)
		}
		t.Log(string(buf[:n]))
		i++
		if i == 10 {
			cancel()
		}
	}

	if i != 10 {
		t.Fatalf("expected 10 events, got %d", i)
	}
}

func TestWebSocketEcho(t *testing.T) {
	t.Parallel()

	if strings.HasPrefix(testConfig.Proxy, "https://") {
		t.Skip("proxy: unknown scheme: https")
	}

	e := Expect(t, testConfig.HTTPBin)

	ws := e.GET("/ws/echo").WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	ws.Subprotocol().IsEmpty()
	for i := 0; i < 100; i++ {
		ws.WriteText(fmt.Sprintf("hello %d", i)).Expect().TextMessage().Body().IsEqual(fmt.Sprintf("hello %d", i))
	}
	ws.CloseWithText("bye").Expect().CloseMessage().NoContent()
}

// func TestProxyLocalhost(t *testing.T) {
//	hosts := []string{
//		"localhost",
//		"127.0.0.1",
//	}
//
//	for _, h := range hosts {
//		if os.Getenv("FORWARDER_PROXY_LOCALHOST") == "allow" {
//			Expect(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").Expect().Status(http.StatusOK)
//		} else {
//			Expect(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").Expect().Status(http.StatusBadGateway)
//		}
//	}
//}

func TestBadGateway(t *testing.T) {
	t.Parallel()

	hosts := []string{
		"wronghost",
		"httpbin:1",
	}

	for _, scheme := range []string{"http", "https"} {
		for _, h := range hosts {
			Expect(t, scheme+"://"+h).GET("/status/200").Expect().Status(http.StatusBadGateway)
		}
	}
}

func TestGoogleCom(t *testing.T) {
	e := Expect(t, "http://www.google.com")
	e.HEAD("/").Expect().Status(http.StatusOK)
}

func TestSC2450(t *testing.T) {
	if testConfig.SC2450 == "" {
		t.Skip("sc2450 not set")
	}

	e := Expect(t, testConfig.SC2450)
	e.HEAD("/").Expect().Status(http.StatusOK)
	e.GET("/").Expect().Status(http.StatusOK).Body().IsEqual(`{"android":{"min_version":"4.0.0"},"ios":{"min_version":"4.0.0"}}`)
}
