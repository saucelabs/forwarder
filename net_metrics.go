// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type listenerMetrics struct {
	accepted  prometheus.Counter
	errors    prometheus.Counter
	tlsErrors prometheus.Counter
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
		tlsErrors: f.NewCounter(prometheus.CounterOpts{
			Name:      "listener_tls_errors_total",
			Namespace: namespace,
			Help:      "Number of TLS handshake errors",
		}),
	}
}

func (m *listenerMetrics) accept() {
	m.accepted.Inc()
}

func (m *listenerMetrics) error() {
	m.errors.Inc()
}

func (m *listenerMetrics) tlsError() {
	m.tlsErrors.Inc()
}
