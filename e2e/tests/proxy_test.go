// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/saucelabs/forwarder/e2e/forwarder"
	"github.com/saucelabs/forwarder/utils/httpexpect"
)

func TestProxyStatusCode(t *testing.T) {
	t.Parallel()
	// List of all valid status codes plus some non-standard ones.
	// See https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
	validStatusCodes := []int{
		// 1xx status codes are different from the rest (e.g. switching protocols is only defined in HTTP/1.1), so we skip them for now.
		// 100, 101, 102, 103, 122,
		200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
		300, 301, 302, 303, 304, 305, 306, 307, 308,
		400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 421, 422, 423, 424, 425, 426, 428, 429, 431, 451,
		500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511, 599,
	}

	methods := []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}

	c := newClient(t, httpbin)
	for i := range validStatusCodes {
		code := validStatusCodes[i]
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			t.Parallel()
			for _, m := range methods {
				c.Request(m, fmt.Sprintf("/status/%d", code)).ExpectStatus(code)
			}
		})
	}
}

func TestProxyBasicAuth(t *testing.T) {
	c := newClient(t, httpbin)
	t.Run("ok", func(t *testing.T) {
		c.GET("/basic-auth/user/passwd", func(r *http.Request) {
			r.SetBasicAuth("user", "passwd")
		}).ExpectStatus(http.StatusOK)
	})
	t.Run("nok", func(t *testing.T) {
		c.GET("/basic-auth/user/passwd").ExpectStatus(http.StatusUnauthorized)
	})
}

func TestProxyAuthRequired(t *testing.T) {
	if basicAuth == "" {
		t.Skip("basic auth not set")
	}
	newClient(t, httpbin, func(tr *http.Transport) {
		p := tr.Proxy
		tr.Proxy = func(req *http.Request) (u *url.URL, err error) {
			u, err = p(req)
			if u != nil {
				u.User = nil
			}
			return
		}
	}).GET("/status/200").ExpectStatus(http.StatusProxyAuthRequired)
}

func TestProxyBytesStreamN(t *testing.T) {
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
		workers = 2 * runtime.NumCPU()
		wg      sync.WaitGroup
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := newClient(t, httpbin)
			for _, p := range rand.Perm(len(sizes)) {
				size := sizes[p]
				c.GET(fmt.Sprintf("/stream-bytes/%d", size)).ExpectStatus(http.StatusOK).ExpectBodySize(size)
			}
		}()
	}

	wg.Wait()
}

func TestProxyServerSentEvents(t *testing.T) {
	tr := newTransport(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, httpbin+"/events/100", http.NoBody)
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

func TestProxyWebSocket(t *testing.T) {
	if strings.HasPrefix(proxy, "https://") {
		t.Skip("proxy: unknown scheme: https")
	}

	proxyURL, err := httpexpect.NewURLWithBasicAuth(proxy, basicAuth)
	if err != nil {
		t.Fatal(err)
	}

	d := *websocket.DefaultDialer
	d.Proxy = http.ProxyURL(proxyURL)
	if tlsCfg, err := defaultTLSConfig(); err != nil {
		t.Fatal(err)
	} else {
		d.TLSClientConfig = tlsCfg
	}

	var u string
	if p, _, _ := strings.Cut(httpbin, ":"); p == "https" {
		u = "wss://httpbin:8080/ws/echo"
	} else {
		u = "ws://httpbin:8080/ws/echo"
	}
	conn, resp, err := d.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	subprotocol := resp.Header.Get("Sec-WebSocket-Protocol")
	if subprotocol != "" {
		t.Fatalf("Subprotocol: %s\n", subprotocol)
	}

	for i := 0; i < 100; i++ {
		message := fmt.Sprintf("hello %d", i)
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			t.Fatalf("Failed to write WebSocket message: %v", err)
		}

		messageType, receivedMessage, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read WebSocket message: %v", err)
		}

		if messageType != websocket.TextMessage {
			t.Fatalf("Expected text message, got message type %d", messageType)
		}

		if string(receivedMessage) != message {
			t.Fatalf("Expected message '%s', got '%s'", message, string(receivedMessage))
		}
	}

	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
	if err != nil {
		t.Fatalf("Failed to close WebSocket: %v", err)
	}

	_, _, err = conn.ReadMessage()
	if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("WebSocket closed unexpectedly: %v", err)
	}
}

func TestProxyBadGateway(t *testing.T) {
	hosts := []string{
		"wronghost",
		"httpbin:1",
	}

	for _, scheme := range []string{"http", "https"} {
		for _, h := range hosts {
			newClient(t, scheme+"://"+h).GET("/status/200").ExpectStatus(http.StatusBadGateway)
		}
	}
}

func TestProxyGoogleCom(t *testing.T) {
	newClient(t, "https://www.google.com").HEAD("/").ExpectStatus(http.StatusOK)
}

func TestProxyUpstream(t *testing.T) {
	if os.Getenv("FORWARDER_PROXY") == "" {
		t.Skip("FORWARDER_PROXY not set")
	}
	if os.Getenv("HTTPBIN_PROTOCOL") != "http" {
		t.Skip("HTTPBIN_PROTOCOL not set to http")
	}

	res := string(newClient(t, httpbin).GET("/header/via/").ExpectStatus(http.StatusNotFound).Body)
	_, viaHeader, ok := strings.Cut(res, "=")
	if !ok {
		t.Fatalf("unexpected response: %q", res)
	}

	l := len("1.1 ")
	filter := func(s string) string {
		i := strings.LastIndex(s, "-")
		if i < l {
			t.Errorf("unexpected via header: %q", s)
			return ""
		}
		return s[l:i]
	}

	var success bool
	for _, via := range strings.Split(viaHeader, ", ") {
		if forwarder.UpstreamProxyServiceName == filter(via) {
			success = true
		}
	}

	if !success {
		t.Fatalf("%s via header not found", forwarder.UpstreamProxyServiceName)
	}
}
