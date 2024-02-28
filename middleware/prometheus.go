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

// Prometheus is a middleware that collects metrics about the HTTP requests and responses.
// Unlike the promhttp.InstrumentHandler* chaining, this middleware creates only one delegator per request.
// It partitions the metrics by HTTP status code, HTTP method, destination host name and source IP.
type Prometheus struct {
	requestsInFlight *prometheus.GaugeVec
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestSize      *prometheus.HistogramVec
	responseSize     *prometheus.HistogramVec
}

func NewPrometheus(r prometheus.Registerer, namespace string) *Prometheus {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)
	l := []string{"code", "method"}

	return &Prometheus{
		requestsInFlight: f.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_in_flight",
			Help:      "Current number of HTTP requests being served.",
		}, []string{"method"}),
		requestsTotal: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests processed.",
		}, l),
		requestDuration: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, l),
		requestSize: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
			Buckets:   sizeBuckets,
		}, l),
		responseSize: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
			Buckets:   sizeBuckets,
		}, l),
	}
}

func (p *Prometheus) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.requestsInFlight.WithLabelValues(r.Method).Inc()

		d := newDelegator(w, nil)

		r.Body = bodyCounter(r.Body)

		start := time.Now()
		h.ServeHTTP(d, r)
		elapsed := time.Since(start).Seconds()

		lv := [2]string{strconv.Itoa(d.Status()), r.Method}

		p.requestsTotal.WithLabelValues(lv[:]...).Inc()
		p.requestDuration.WithLabelValues(lv[:]...).Observe(elapsed)

		reqSize := int64(0)
		if c, ok := r.Body.(counter); ok {
			reqSize = c.Count()
		}

		p.requestsInFlight.WithLabelValues(r.Method).Dec()
		p.requestSize.WithLabelValues(lv[:]...).Observe(float64(reqSize))
		p.responseSize.WithLabelValues(lv[:]...).Observe(float64(d.Written()))
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
	p.requestsInFlight.WithLabelValues(req.Method).Inc()
}

func (p *Prometheus) WroteResponse(res *http.Response) {
	elapsed := martian.ContextDuration(res.Request.Context()).Seconds()

	req := res.Request
	lv := [2]string{strconv.Itoa(res.StatusCode), req.Method}

	p.requestsInFlight.WithLabelValues(req.Method).Dec()
	p.requestsTotal.WithLabelValues(lv[:]...).Inc()
	p.requestDuration.WithLabelValues(lv[:]...).Observe(elapsed)

	reqSize := int64(0)
	if c, ok := req.Body.(counter); ok {
		reqSize = c.Count()
	}
	p.requestSize.WithLabelValues(lv[:]...).Observe(float64(reqSize))

	resSize := int64(0)
	if c, ok := res.Body.(counter); ok {
		resSize = c.Count()
	}
	p.responseSize.WithLabelValues(lv[:]...).Observe(float64(resSize))
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
