// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder/log"
	"github.com/saucelabs/forwarder/utils/certutil"
	"github.com/saucelabs/forwarder/utils/golden"
)

func TestDialRedirectFromHostPortPairs(t *testing.T) {
	tests := []struct {
		name  string
		hp    HostPortPair
		input string
		want  string
	}{
		{
			name: "basic",
			hp: HostPortPair{
				Src: HostPort{"a", "80"},
				Dst: HostPort{"b", "443"},
			},
			input: "a:80",
			want:  "b:443",
		},
		{
			name: "blank src host",
			hp: HostPortPair{
				Src: HostPort{"", "80"},
				Dst: HostPort{"b", "443"},
			},
			input: "a:80",
			want:  "b:443",
		},
		{
			name: "blank src port",
			hp: HostPortPair{
				Src: HostPort{"a", ""},
				Dst: HostPort{"b", "443"},
			},
			input: "a:80",
			want:  "b:443",
		},
		{
			name: "blank dst host",
			hp: HostPortPair{
				Src: HostPort{"a", "80"},
				Dst: HostPort{"", "443"},
			},
			input: "a:80",
			want:  "a:443",
		},
		{
			name: "blank dst port",
			hp: HostPortPair{
				Src: HostPort{"a", "80"},
				Dst: HostPort{"b", ""},
			},
			input: "a:80",
			want:  "b:80",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			_, got := DialRedirectFromHostPortPairs([]HostPortPair{tc.hp})("tcp", tc.input)
			if got != tc.want {
				t.Fatalf("DialRedirectFromHostPortPairs(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDialerRedirect(t *testing.T) {
	l := Listener{
		Address: "localhost:0",
		Log:     log.NopLogger,
	}
	defer l.Close()

	l.listenAndWait(t)
	go l.acceptAndCopy()

	d := NewDialer(&DialConfig{
		DialTimeout: 10 * time.Millisecond,
		RedirectFunc: func(network, address string) (string, string) {
			return "tcp", l.Addr().String()
		},
	})

	ctx := context.Background()
	conn, err := d.DialContext(ctx, "tcp", "foo")
	if err != nil {
		t.Fatalf("d.DialContext(): got %v, want no error", err)
	}
	defer conn.Close()

	fmt.Fprintf(conn, "Hello, World!\n")
	buf := make([]byte, 20)
	if n, err := conn.Read(buf); err != nil {
		t.Fatal(err)
	} else {
		buf = buf[:n]
	}
	if got, want := string(buf), "Hello, World!\n"; got != want {
		t.Fatalf("conn.Read(): got %q, want %q", got, want)
	}
}

func TestDialerMetrics(t *testing.T) {
	l := Listener{
		Address: "localhost:0",
		Log:     log.NopLogger,
	}
	defer l.Close()

	l.listenAndWait(t)
	go l.acceptAndCopy()

	r := prometheus.NewRegistry()
	d := NewDialer(&DialConfig{
		DialTimeout: 10 * time.Millisecond,
		PromConfig: PromConfig{
			PromNamespace: "test",
			PromRegistry:  r,
		},
	})

	ctx := context.Background()
	for range 10 {
		conn, err := d.DialContext(ctx, "tcp", l.Addr().String())
		if err != nil {
			t.Fatalf("d.DialContext(): got %v, want no error", err)
		}
		fmt.Fprintf(conn, "Hello, World!\n")
		if _, err := conn.Read(make([]byte, 1)); err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}

	golden.DiffPrometheusMetrics(t, r)
}

func TestDialerMetricsErrors(t *testing.T) {
	r := prometheus.NewRegistry()
	d := NewDialer(&DialConfig{
		DialTimeout: 10 * time.Millisecond,
		PromConfig: PromConfig{
			PromNamespace: "test",
			PromRegistry:  r,
		},
	})

	ctx := context.Background()
	_, err := d.DialContext(ctx, "tcp", "localhost:0")
	if err == nil {
		t.Fatal("d.DialContext(): got no error, want error")
	}

	golden.DiffPrometheusMetrics(t, r)
}

func (l *Listener) listenAndWait(t *testing.T) {
	t.Helper()

	if err := l.Listen(); err != nil {
		t.Fatal(err)
	}
	for {
		if l.Addr() != nil {
			break
		}
	}
}

func (l *Listener) acceptAndCopy() {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go func() {
			io.Copy(conn, conn)
			conn.Close()
		}()
	}
}

func TestListenerListenOnce(t *testing.T) {
	l := Listener{
		Address: "localhost:0",
		Log:     log.NopLogger,
	}
	defer l.Close()

	l.listenAndWait(t)

	if err := l.Listen(); err == nil {
		t.Fatal("l.Listen(): got no error, want error")
	}
}

func TestListenerMetricsAccepted(t *testing.T) {
	r := prometheus.NewRegistry()
	l := Listener{
		Address: "localhost:0",
		Log:     log.NopLogger,
		PromConfig: PromConfig{
			PromNamespace: "test",
			PromRegistry:  r,
		},
	}
	defer l.Close()

	l.listenAndWait(t)
	go l.acceptAndCopy()

	for range 10 {
		conn, err := net.Dial("tcp", l.Addr().String())
		if err != nil {
			t.Fatalf("net.Dial(): got %v, want no error", err)
		}
		fmt.Fprintf(conn, "Hello, World!\n")
		if _, err := conn.Read(make([]byte, 1)); err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}

	golden.DiffPrometheusMetrics(t, r)
}

func TestListenerMetricsAcceptedWithTLS(t *testing.T) {
	r := prometheus.NewRegistry()
	l := Listener{
		Address:   "localhost:0",
		Log:       log.NopLogger,
		TLSConfig: selfSingedCert(),
		PromConfig: PromConfig{
			PromNamespace: "test",
			PromRegistry:  r,
		},
	}
	defer l.Close()

	l.listenAndWait(t)
	go l.acceptAndCopy()

	for range 10 {
		conn, err := net.Dial("tcp", l.Addr().String())
		if err != nil {
			t.Fatalf("net.Dial(): got %v, want no error", err)
		}
		conn = tls.Client(conn, &tls.Config{InsecureSkipVerify: true})
		fmt.Fprintf(conn, "Hello, World!\n")
		if _, err := conn.Read(make([]byte, 1)); err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}

	golden.DiffPrometheusMetrics(t, r)
}

func TestListenerMetricsClosed(t *testing.T) {
	r := prometheus.NewRegistry()
	l := Listener{
		Address: "localhost:0",
		Log:     log.NopLogger,
		PromConfig: PromConfig{
			PromNamespace: "test",
			PromRegistry:  r,
		},
	}
	defer l.Close()

	l.listenAndWait(t)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		conn.Close()
		conn.Close() // Close twice, the second time should not be counted.
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	conn.Close()

	golden.DiffPrometheusMetrics(t, r)
}

type errListener struct {
	net.Listener
}

func (l errListener) Accept() (net.Conn, error) {
	return nil, errors.New("accept error")
}

func TestListenerMetricsErrors(t *testing.T) {
	r := prometheus.NewRegistry()
	l := Listener{
		Address: "localhost:0",
		Log:     log.NopLogger,
		PromConfig: PromConfig{
			PromNamespace: "test",
			PromRegistry:  r,
		},
	}
	defer l.Close()

	l.listenAndWait(t)
	l.listener = errListener{l.listener}

	go l.acceptAndCopy()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	conn.Close()

	golden.DiffPrometheusMetrics(t, r)
}

func TestListenerTLSHandshakeTimeout(t *testing.T) {
	r := prometheus.NewRegistry()
	l := Listener{
		Address:             "localhost:0",
		Log:                 log.NopLogger,
		TLSConfig:           selfSingedCert(),
		TLSHandshakeTimeout: 100 * time.Millisecond,
		PromConfig: PromConfig{
			PromNamespace: "test",
			PromRegistry:  r,
		},
	}
	defer l.Close()

	l.listenAndWait(t)
	go l.acceptAndCopy()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatalf("net.Dial(): got %v, want no error", err)
	}
	defer conn.Close()

	time.Sleep(l.TLSHandshakeTimeout * 2)

	golden.DiffPrometheusMetrics(t, r)
}

func selfSingedCert() *tls.Config {
	ssc := certutil.ECDSASelfSignedCert()
	ssc.Hosts = append(ssc.Hosts, "localhost")
	cert, err := ssc.Gen()
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}
