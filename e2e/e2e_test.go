// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package e2e

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

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

	c := newClient(t, httpbin)
	for i := range validStatusCodes {
		code := validStatusCodes[i]
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			t.Parallel()
			c.GET(fmt.Sprintf("/status/%d", code)).ExpectStatus(code)
			c.HEAD(fmt.Sprintf("/status/%d", code)).ExpectStatus(code)
		})
	}
}

func TestAuth(t *testing.T) {
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

func TestProxyAuth(t *testing.T) {
	if os.Getenv("FORWARDER_BASIC_AUTH") == "" {
		t.Skip("FORWARDER_BASIC_AUTH not set")
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

func TestStreamBytes(t *testing.T) {
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

func TestServerSentEvents(t *testing.T) {
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

func TestWebSocketEcho(t *testing.T) {
	if strings.HasPrefix(proxy, "https://") {
		t.Skip("proxy: unknown scheme: https")
	}

	proxyURL := newProxyURL(t)
	d := *websocket.DefaultDialer
	d.Proxy = http.ProxyURL(proxyURL)
	d.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
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
		t.Errorf("Expected status %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	subprotocol := resp.Header.Get("Sec-WebSocket-Protocol")
	if subprotocol != "" {
		t.Errorf("Subprotocol: %s\n", subprotocol)
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
			t.Errorf("Expected text message, got message type %d", messageType)
		}

		if string(receivedMessage) != message {
			t.Errorf("Expected message '%s', got '%s'", message, string(receivedMessage))
		}
	}

	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
	if err != nil {
		t.Fatalf("Failed to close WebSocket: %v", err)
	}

	_, _, err = conn.ReadMessage()
	if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Errorf("WebSocket closed unexpectedly: %v", err)
	}
}

func TestProxyLocalhost(t *testing.T) {
	hosts := []string{
		"localhost",
		"127.0.0.1",
	}

	for _, h := range hosts {
		if os.Getenv("FORWARDER_PROXY_LOCALHOST") == "allow" {
			newClient(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").ExpectStatus(http.StatusOK)
		} else {
			newClient(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").ExpectStatus(http.StatusBadGateway)
		}
	}
}

func TestBadGateway(t *testing.T) {
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

func TestGoogleCom(t *testing.T) {
	newClient(t, "https://www.google.com").HEAD("/").ExpectStatus(http.StatusOK)
}

func TestSC2450(t *testing.T) {
	if os.Getenv("FORWARDER_SC2450") == "" {
		t.Skip("FORWARDER_SC2450 not set")
	}

	c := newClient(t, "http://sc-2450:8307")
	c.HEAD("/").ExpectStatus(http.StatusOK)
	c.GET("/").ExpectStatus(http.StatusOK).ExpectBodyContent(`{"android":{"min_version":"4.0.0"},"ios":{"min_version":"4.0.0"}}`)
}

func TestHeaderMods(t *testing.T) {
	if os.Getenv("FORWARDER_TEST_HEADERS") == "" {
		t.Skip("FORWARDER_TEST_HEADERS not set")
	}

	c := newClient(t, httpbin)
	c.GET("/header/test-add/test-value").ExpectStatus(http.StatusOK)
	c.GET("/header/test-empty/", func(r *http.Request) {
		r.Header.Set("test-empty", "not-empty")
	}).ExpectStatus(http.StatusOK)
	c.GET("/header/test-rm/value-1", func(r *http.Request) {
		r.Header.Set("test-rm", "value-1")
	}).ExpectStatus(http.StatusNotFound)
	c.GET("/header/rm-prefix/value-2", func(r *http.Request) {
		r.Header.Set("rm-prefix", "value-2")
	}).ExpectStatus(http.StatusNotFound)
}

func TestHeaderRespMods(t *testing.T) {
	if os.Getenv("FORWARDER_TEST_RESPONSE_HEADERS") == "" {
		t.Skip("FORWARDER_TEST_RESPONSE_HEADERS not set")
	}

	c := newClient(t, httpbin)
	c.GET("/status/200").ExpectStatus(http.StatusOK).ExpectHeader("test-resp-add", "test-resp-value")
	c.GET("/header/test-resp-empty/not-empty", func(r *http.Request) {
		r.Header.Set("test-resp-empty", "not-empty")
	}).ExpectStatus(http.StatusOK).ExpectHeader("test-resp-empty", "")
	c.GET("/header/test-resp-rm/value-3", func(r *http.Request) {
		r.Header.Set("test-resp-rm", "value-3")
	}).ExpectStatus(http.StatusOK).ExpectHeader("test-resp-rm", "")
	c.GET("/header/resp-rm-prefix/value-4", func(r *http.Request) {
		r.Header.Set("resp-rm-prefix", "value-4")
	}).ExpectStatus(http.StatusOK).ExpectHeader("resp-rm-prefix", "")
}
