// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
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
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
		t.Run(strconv.Itoa(code), func(t *testing.T) {
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
		for range n {
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

	for range workers {
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

	for i := range 100 {
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

	expectedErrorMessage := "forwarder dial tcp"
	if os.Getenv("FORWARDER_PROXY") != "" {
		// Make sure the error originates from the upstream proxy.
		expectedErrorMessage = forwarder.UpstreamProxyServiceName + " dial tcp"
	} else if os.Getenv("FORWARDER_PAC") != "" {
		// In case of PAC, we can never now the exact proxy that has been used.
		expectedErrorMessage = "dial tcp"
	}

	enableInsecureSkipVerifyIfCAIsGenerated := func(tr *http.Transport) {
		if os.Getenv("SETUP") == "flag-mitm-genca" {
			tr.TLSClientConfig.InsecureSkipVerify = true
		}
	}

	var cres *http.Response
	saveProxyConnectResponse := func(tr *http.Transport) {
		tr.OnProxyConnectResponse = func(ctx context.Context, proxyURL *url.URL, connectReq *http.Request, connectRes *http.Response) error {
			cres = connectRes
			return nil
		}
	}

	for _, scheme := range []string{"http", "https"} {
		for _, h := range hosts {
			t.Run(scheme+"_"+h, func(t *testing.T) {
				c := newClient(t, scheme+"://"+h,
					enableInsecureSkipVerifyIfCAIsGenerated,
					saveProxyConnectResponse)

				res := c.GET("/status/200")
				res.ExpectStatus(http.StatusBadGateway)

				// Fallback to cres if no response was received - this is to work around Go behavior.
				if len(res.Header) == 0 {
					res = c.MakeResponse(cres)
				}

				if msg := res.Header.Get("X-Forwarder-Error"); !strings.Contains(msg, expectedErrorMessage) {
					t.Fatalf("Expected error message to contain %q, got %q", expectedErrorMessage, msg)
				}
			})
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

	viaHeader := newClient(t, httpbin).GET("/headers/").Header["Via"]
	var success bool
	for _, via := range viaHeader {
		if strings.Contains(via, forwarder.UpstreamProxyServiceName) {
			success = true
		}
	}

	if !success {
		t.Fatalf("%s via header not found", forwarder.UpstreamProxyServiceName)
	}
}

func TestProxyReuseConnection(t *testing.T) {
	c := newClient(t, "http://wronghost", func(tr *http.Transport) {
		var (
			d      net.Dialer
			dialed atomic.Bool
		)
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if dialed.CompareAndSwap(false, true) {
				return d.DialContext(ctx, network, addr)
			}
			return nil, errors.New("only one dial is allowed")
		}
	})

	for range 2 {
		r, w := io.Pipe()

		written := make(chan struct{})
		go func() {
			zeros := make([]byte, 1024*1024)
			if _, err := w.Write(zeros); err != nil {
				t.Error(err)
			}
			if err := w.Close(); err != nil {
				t.Error(err)
			}
			close(written)
		}()

		res := c.GET("/", func(req *http.Request) {
			req.Body = r
		})
		if res.StatusCode != http.StatusBadGateway {
			t.Fatalf("Expected status %d, got %d", http.StatusBadGateway, res.StatusCode)
		}

		select {
		case <-written:
		case <-time.After(10 * time.Second):
			t.Fatal("timed-out waiting for body read")
		}
	}
}
