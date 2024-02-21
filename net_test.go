// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
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
		Address:       "localhost:0",
		Log:           log.NopLogger,
		PromNamespace: "test",
		PromRegistry:  r,
	}
	defer l.Close()

	l.listenAndWait(t)
	go l.acceptAndCopy()

	for i := 0; i < 10; i++ {
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

	// Wait for the metrics to be updated.
	// Somehow, the metrics are not updated immediately.
	time.Sleep(10 * time.Millisecond)

	golden.DiffPrometheusMetrics(t, r)
}

func TestListenerMetricsAcceptedWithTLS(t *testing.T) {
	r := prometheus.NewRegistry()
	l := Listener{
		Address:       "localhost:0",
		Log:           log.NopLogger,
		TLSConfig:     selfSingedCert(),
		PromNamespace: "test",
		PromRegistry:  r,
	}
	defer l.Close()

	l.listenAndWait(t)
	go l.acceptAndCopy()

	for i := 0; i < 10; i++ {
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
		Address:       "localhost:0",
		Log:           log.NopLogger,
		PromNamespace: "test",
		PromRegistry:  r,
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
		Address:       "localhost:0",
		Log:           log.NopLogger,
		PromNamespace: "test",
		PromRegistry:  r,
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
		PromNamespace:       "test",
		PromRegistry:        r,
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
