// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net"
	"slices"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type dialerMetrics struct {
	errors *prometheus.CounterVec
	dialed *prometheus.CounterVec
	active *prometheus.GaugeVec
}

func newDialerMetrics(r prometheus.Registerer, namespace string) *dialerMetrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)
	l := []string{"host"}

	return &dialerMetrics{
		errors: f.NewCounterVec(prometheus.CounterOpts{
			Name:      "dialer_errors_total",
			Namespace: namespace,
			Help:      "Number of errors dialing connections",
		}, l),
		dialed: f.NewCounterVec(prometheus.CounterOpts{
			Name:      "dialer_cx_total",
			Namespace: namespace,
			Help:      "Number of dialed connections",
		}, l),
		active: f.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "dialer_cx_active",
			Namespace: namespace,
			Help:      "Number of active connections",
		}, l),
	}
}

func (m *dialerMetrics) error(addr string) {
	m.errors.WithLabelValues(addr2Host(addr)).Inc()
}

func (m *dialerMetrics) dial(addr string) {
	host := addr2Host(addr)
	m.dialed.WithLabelValues(host).Inc()
	m.active.WithLabelValues(host).Inc()
}

func (m *dialerMetrics) close(addr string) {
	m.active.WithLabelValues(addr2Host(addr)).Dec()
}

func addr2Host(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "unknown"
	}

	commonLocalhostNames := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"::",
	}
	if slices.Contains(commonLocalhostNames, host) {
		return "localhost"
	}

	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsUnspecified()) {
		return "localhost"
	}

	return host
}

type listenerMetrics struct {
	accepted prometheus.Counter
	errors   prometheus.Counter
	closed   prometheus.Counter
}

func newListenerMetrics(r prometheus.Registerer, namespace string) *listenerMetrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)

	return &listenerMetrics{
		accepted: f.NewCounter(prometheus.CounterOpts{
			Name:      "listener_accepted_total",
			Namespace: namespace,
			Help:      "Number of accepted connections",
		}),
		errors: f.NewCounter(prometheus.CounterOpts{
			Name:      "listener_errors_total",
			Namespace: namespace,
			Help:      "Number of listener errors when accepting connections",
		}),
		closed: f.NewCounter(prometheus.CounterOpts{
			Name:      "listener_closed_total",
			Namespace: namespace,
			Help:      "Number of closed connections",
		}),
	}
}

func (m *listenerMetrics) accept() {
	m.accepted.Inc()
}

func (m *listenerMetrics) error() {
	m.errors.Inc()
}

func (m *listenerMetrics) close() {
	m.closed.Inc()
}

func newListenerMetricsWithNameFunc(r prometheus.Registerer, namespace string) func(name string) *listenerMetrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)

	accepted := f.NewCounterVec(prometheus.CounterOpts{
		Name:      "listener_accepted_total",
		Namespace: namespace,
		Help:      "Number of accepted connections",
	}, []string{"name"})
	errors := f.NewCounterVec(prometheus.CounterOpts{
		Name:      "listener_errors_total",
		Namespace: namespace,
		Help:      "Number of listener errors when accepting connections",
	}, []string{"name"})
	closed := f.NewCounterVec(prometheus.CounterOpts{
		Name:      "listener_closed_total",
		Namespace: namespace,
		Help:      "Number of closed connections",
	}, []string{"name"})

	return func(name string) *listenerMetrics {
		return &listenerMetrics{
			accepted: accepted.WithLabelValues(name),
			errors:   errors.WithLabelValues(name),
			closed:   closed.WithLabelValues(name),
		}
	}
}
