// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/saucelabs/forwarder/internal/martian"
)

const (
	_          = iota // ignore first value by assigning to blank identifier
	kb float64 = 1 << (10 * iota)
	mb
)

var sizeBuckets = []float64{ //nolint:gochecknoglobals // this is a global variable by design
	1 * kb,
	2 * kb,
	5 * kb,
	10 * kb,
	100 * kb,
	500 * kb,
	1 * mb,
	2.5 * mb,
	5 * mb,
	10 * mb,
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
	requestDuration  *prometheus.HistogramVec
	requestSize      *prometheus.HistogramVec
	responseSize     *prometheus.HistogramVec

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

	p.requestDuration = f.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "The HTTP request latencies in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, labelsWithStatus)

	p.requestSize = f.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_size_bytes",
		Help:      "The HTTP request sizes in bytes.",
		Buckets:   sizeBuckets,
	}, labelsWithStatus)

	p.responseSize = f.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_response_size_bytes",
		Help:      "The HTTP response sizes in bytes.",
		Buckets:   sizeBuckets,
	}, labelsWithStatus)

	return p
}

func (p *Prometheus) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		labels := p.labels(r)

		p.requestsInFlight.WithLabelValues(labels...).Inc()

		d := newDelegator(w, nil)

		r.Body = bodyCounter(r.Body)

		start := time.Now()
		h.ServeHTTP(d, r)
		elapsed := time.Since(start).Seconds()

		statusLabel := strconv.Itoa(d.Status())
		labelsWithStatus := append([]string{statusLabel}, labels...)

		p.requestsTotal.WithLabelValues(labelsWithStatus...).Inc()
		p.requestDuration.WithLabelValues(labelsWithStatus...).Observe(elapsed)

		reqSize := int64(0)
		if c, ok := r.Body.(counter); ok {
			reqSize = c.Count()
		}

		p.requestsInFlight.WithLabelValues(labels...).Dec()
		p.requestSize.WithLabelValues(labelsWithStatus...).Observe(float64(reqSize))
		p.responseSize.WithLabelValues(labelsWithStatus...).Observe(float64(d.Written()))
	})
}

func (p *Prometheus) ModifyRequest(req *http.Request) error {
	req.Body = bodyCounter(req.Body)
	return nil
}

func (p *Prometheus) ModifyResponse(res *http.Response) error {
	res.Body = bodyCounter(res.Body)
	return nil
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

	reqSize := int64(0)
	if c, ok := req.Body.(counter); ok {
		reqSize = c.Count()
	}
	p.requestSize.WithLabelValues(labelsWithStatus...).Observe(float64(reqSize))

	resSize := int64(0)
	if c, ok := res.Body.(counter); ok {
		resSize = c.Count()
	}
	p.responseSize.WithLabelValues(labelsWithStatus...).Observe(float64(resSize))
}

func (p *Prometheus) labels(req *http.Request) []string {
	labels := []string{req.Method}
	if p.label != "" {
		labels = append(labels, p.labeler(req))
	}
	return labels
}

type counter interface {
	Count() int64
}

func bodyCounter(b io.ReadCloser) io.ReadCloser {
	if b == nil || b == http.NoBody {
		return b
	}

	if _, ok := b.(io.ReadWriteCloser); ok {
		return &rwcBody{body{ReadCloser: b}}
	}

	return &body{ReadCloser: b}
}

type body struct {
	io.ReadCloser
	n int64
}

func (b *body) Count() int64 {
	return b.n
}

func (b *body) Read(p []byte) (n int, err error) {
	n, err = b.ReadCloser.Read(p)
	b.n += int64(n)
	return
}

type rwcBody struct {
	body
}

func (b *rwcBody) Write(p []byte) (int, error) {
	return b.ReadCloser.(io.ReadWriteCloser).Write(p) //nolint:forcetypeassert // We know it's a ReadWriteCloser.
}
