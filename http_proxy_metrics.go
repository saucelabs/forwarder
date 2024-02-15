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

type httpProxyMetrics struct {
	errors *prometheus.CounterVec
}

func newHTTPProxyMetrics(r prometheus.Registerer, namespace string) *httpProxyMetrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)

	return &httpProxyMetrics{
		errors: f.NewCounterVec(prometheus.CounterOpts{
			Name:      "proxy_errors_total",
			Namespace: namespace,
			Help:      "Number of proxy errors",
		}, []string{"reason"}),
	}
}

func (m *httpProxyMetrics) error(reason string) {
	m.errors.WithLabelValues(reason).Inc()
}
