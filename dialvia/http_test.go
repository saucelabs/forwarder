// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dialvia

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/internal/martian/proxyutil"
	"golang.org/x/net/context"
)

func TestHTTPProxyDialerDialContext(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	d := HTTPProxy(
		(&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		&url.URL{Scheme: "http", Host: l.Addr().String()},
	)

	ctx := context.Background()

	t.Run("status 200", func(t *testing.T) {
		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				pbr := bufio.NewReader(conn)
				req, err := http.ReadRequest(pbr)
				if err != nil {
					return err
				}
				return proxyutil.NewResponse(200, nil, req).Write(conn)
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err != nil {
			t.Fatal(err)
		}
		if conn == nil {
			t.Fatal("conn is nil")
		}

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("status 404", func(t *testing.T) {
		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				pbr := bufio.NewReader(conn)
				req, err := http.ReadRequest(pbr)
				if err != nil {
					return err
				}
				return proxyutil.NewResponse(404, nil, req).Write(conn)
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err == nil {
			t.Fatal("err is nil")
		}
		t.Log(err)
		if conn != nil {
			t.Fatal("conn is not nil")
		}

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CONNECT headers from GetProxyConnectHeaders", func(t *testing.T) {
		d.GetProxyConnectHeader = func(_ context.Context, proxyURL *url.URL, _ string) (http.Header, error) {
			authHeader := make(http.Header, 1)
			authHeader.Set("Proxy-Authorization", "TEST-PROXY-AUTHORIZATION")
			return authHeader, nil
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				pbr := bufio.NewReader(conn)
				req, err := http.ReadRequest(pbr)
				if err != nil {
					return err
				}

				if req.Method != http.MethodConnect {
					return fmt.Errorf("HTTP CONNECT method expected")
				}

				if req.Header.Get("Proxy-Authorization") != "TEST-PROXY-AUTHORIZATION" {
					return fmt.Errorf("Proxy-Authorization header expected but not present")
				}

				return proxyutil.NewResponse(404, nil, req).Write(conn)
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err == nil {
			t.Fatal("err is nil")
		}
		t.Log(err)
		if conn != nil {
			t.Fatal("conn is not nil")
		}

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("response with data", func(t *testing.T) {
		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				pbr := bufio.NewReader(conn)
				req, err := http.ReadRequest(pbr)
				if err != nil {
					return err
				}

				// Write a response with data.
				pbw := bufio.NewWriter(conn)
				proxyutil.NewResponse(200, nil, req).Write(pbw)
				for i := range 100 {
					fmt.Fprintf(pbw, "hello %d\n", i)
				}
				return pbw.Flush()
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err != nil {
			t.Fatal(err)
		}
		if conn == nil {
			t.Fatal("conn is nil")
		}

		n := 0
		s := bufio.NewScanner(conn)
		for s.Scan() {
			n++
		}
		if err := s.Err(); err != nil {
			t.Fatal(err)
		}
		if n != 100 {
			t.Fatalf("n=%d, want 100", n)
		}
		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("conn closed", func(t *testing.T) {
		errCh := make(chan error, 1)
		go func() {
			errCh <- serveOne(l, func(conn net.Conn) error {
				conn.Close()
				return nil
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err == nil {
			t.Fatal("err is nil")
		}
		t.Log(err)
		if conn != nil {
			t.Fatal("conn is not nil")
		}

		if err := <-errCh; err != nil {
			t.Fatal(err)
		}
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		done := make(chan struct{})
		go func() {
			serveOne(l, func(conn net.Conn) error {
				cancel()
				<-done
				return nil
			})
		}()

		conn, err := d.DialContext(ctx, "tcp", "foobar.com:80")
		if err == nil {
			t.Fatal("err is nil")
		}
		t.Log(err)
		if conn != nil {
			t.Fatal("conn is not nil")
		}

		done <- struct{}{}
	})
}

func serveOne(l net.Listener, h func(conn net.Conn) error) error {
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	return h(conn)
}

func BenchmarkHTTPProxyDialer_DialContextR(b *testing.B) {
	b.Run("status 200", func(b *testing.B) {
		var buf bytes.Buffer
		resp := proxyutil.NewResponse(200, nil, nil)
		resp.Write(&buf)
		benchmarkHTTPProxyDialerDialContextR(b, buf.Bytes())
	})

	b.Run("status 403", func(b *testing.B) {
		var buf bytes.Buffer
		resp := proxyutil.NewResponse(403, bytes.NewBufferString("proxying is denied to host \"foobar\"\nproxying denied"), nil)
		resp.Header.Set("Content-Type", "text/plain; charset=utf-8")
		resp.Header.Set("X-Forwarder-Error", "proxying denied")
		resp.Write(&buf)
		benchmarkHTTPProxyDialerDialContextR(b, buf.Bytes())
	})
}

func benchmarkHTTPProxyDialerDialContextR(b *testing.B, resp []byte) {
	b.Helper()

	mockDialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		c0, c1 := net.Pipe()
		go func() {
			defer c0.Close()
			if _, err := io.CopyN(io.Discard, c0, 55); err != nil {
				b.Errorf("CopyN: %v", err)
			}
			if _, err := c0.Write(resp); err != nil {
				b.Errorf("Write: %v", err)
			}
		}()
		return c1, nil
	}

	d := HTTPProxy(mockDialer, &url.URL{Scheme: "http", Host: "foobar.com:80"})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := d.DialContextR(ctx, "tcp", "foobar.com:80")
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}
