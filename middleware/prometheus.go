// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/saucelabs/forwarder/internal/martian"
)

var objectives = map[float64]float64{
	0.5:  0.01,  // Median (50th percentile) with ±1% error
	0.9:  0.01,  // 90th percentile with ±1% error
	0.99: 0.001, // 99th percentile with ±0.1% error
}

type PrometheusOpt func(*Prometheus)

type PrometheusLabeler func(*http.Request) string

func WithCustomLabeler(label string, labeler PrometheusLabeler) PrometheusOpt {
	return func(p *Prometheus) {
		p.label = label
		p.labeler = labeler
	}
}

// Prometheus is a middleware that collects metrics about the HTTP requests and responses.
// Unlike the promhttp.InstrumentHandler* chaining, this middleware creates only one delegator per request.
// It partitions the metrics by HTTP status code, HTTP method, destination host name and source IP.
type Prometheus struct {
	requestsInFlight *prometheus.GaugeVec
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.SummaryVec
	// The following metrics are now removed, revert if needed.
	// requestSize      *prometheus.SummaryVec
	// responseSize     *prometheus.SummaryVec

	label   string
	labeler PrometheusLabeler
}

func NewPrometheus(r prometheus.Registerer, namespace string, opts ...PrometheusOpt) *Prometheus {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)

	p := &Prometheus{}
	for _, opt := range opts {
		opt(p)
	}

	labels := []string{"method"}
	if p.label != "" {
		labels = append(labels, p.label)
	}
	labelsWithStatus := append([]string{"code"}, labels...)

	p.requestsInFlight = f.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "http_requests_in_flight",
		Help:      "Current number of HTTP requests being served.",
	}, labels)

	p.requestsTotal = f.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests processed.",
	}, labelsWithStatus)

	p.requestDuration = f.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  namespace,
		Name:       "http_request_duration_seconds",
		Help:       "The HTTP request latencies in seconds.",
		Objectives: objectives,
	}, labelsWithStatus)

	return p
}

func (p *Prometheus) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labels := p.labels(r)

		p.requestsInFlight.WithLabelValues(labels...).Inc()

		d := newDelegator(w, nil)

		start := time.Now()
		h.ServeHTTP(d, r)
		elapsed := time.Since(start).Seconds()

		statusLabel := strconv.Itoa(d.Status())
		labelsWithStatus := append([]string{statusLabel}, labels...)

		p.requestsTotal.WithLabelValues(labelsWithStatus...).Inc()
		p.requestDuration.WithLabelValues(labelsWithStatus...).Observe(elapsed)

		p.requestsInFlight.WithLabelValues(labels...).Dec()
	})
}

func (p *Prometheus) ReadRequest(req *http.Request) {
	p.requestsInFlight.WithLabelValues(p.labels(req)...).Inc()
}

func (p *Prometheus) WroteResponse(res *http.Response) {
	elapsed := martian.ContextDuration(res.Request.Context()).Seconds()

	req := res.Request

	labels := p.labels(req)
	statusLabel := strconv.Itoa(res.StatusCode)
	labelsWithStatus := append([]string{statusLabel}, labels...)

	p.requestsInFlight.WithLabelValues(labels...).Dec()
	p.requestsTotal.WithLabelValues(labelsWithStatus...).Inc()
	p.requestDuration.WithLabelValues(labelsWithStatus...).Observe(elapsed)
}

func (p *Prometheus) labels(req *http.Request) []string {
	labels := []string{req.Method}
	if p.label != "" {
		labels = append(labels, p.labeler(req))
	}
	return labels
}
