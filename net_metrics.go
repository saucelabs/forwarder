// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
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
	retries *prometheus.CounterVec
	errors  *prometheus.CounterVec
	dialed  *prometheus.CounterVec
	active  *prometheus.GaugeVec
}

func newDialerMetrics(r prometheus.Registerer, namespace string) *dialerMetrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)
	l := []string{"host"}

	return &dialerMetrics{
		retries: f.NewCounterVec(prometheus.CounterOpts{
			Name:      "dialer_retries_total",
			Namespace: namespace,
			Help:      "Number of dial retries",
		}, l),
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

func (m *dialerMetrics) retry(addr string) {
	m.retries.WithLabelValues(addr2Host(addr)).Inc()
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
	errors   prometheus.Counter
	accepted prometheus.Counter
	active   prometheus.Gauge
}

func newListenerMetrics(r prometheus.Registerer, namespace string) *listenerMetrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)

	return &listenerMetrics{
		errors: f.NewCounter(prometheus.CounterOpts{
			Name:      "listener_errors_total",
			Namespace: namespace,
			Help:      "Number of listener errors when accepting connections",
		}),
		accepted: f.NewCounter(prometheus.CounterOpts{
			Name:      "listener_cx_total",
			Namespace: namespace,
			Help:      "Number of accepted connections",
		}),
		active: f.NewGauge(prometheus.GaugeOpts{
			Name:      "listener_cx_active",
			Namespace: namespace,
			Help:      "Number of active connections",
		}),
	}
}

func (m *listenerMetrics) error() {
	m.errors.Inc()
}

func (m *listenerMetrics) accept() {
	m.accepted.Inc()
	m.active.Inc()
}

func (m *listenerMetrics) close() {
	m.active.Dec()
}

func newListenerMetricsWithNameFunc(r prometheus.Registerer, namespace string) func(name string) *listenerMetrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)

	errors := f.NewCounterVec(prometheus.CounterOpts{
		Name:      "listener_errors_total",
		Namespace: namespace,
		Help:      "Number of listener errors when accepting connections",
	}, []string{"name"})
	accepted := f.NewCounterVec(prometheus.CounterOpts{
		Name:      "listener_cx_total",
		Namespace: namespace,
		Help:      "Number of accepted connections",
	}, []string{"name"})
	active := f.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "listener_cx_active",
		Namespace: namespace,
		Help:      "Number of active connections",
	}, []string{"name"})

	return func(name string) *listenerMetrics {
		return &listenerMetrics{
			errors:   errors.WithLabelValues(name),
			accepted: accepted.WithLabelValues(name),
			active:   active.WithLabelValues(name),
		}
	}
}
