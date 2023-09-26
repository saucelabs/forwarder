// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"fmt"
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
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
}

func NewPrometheus(r prometheus.Registerer, namespace string) *Prometheus {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)
	l := []string{"code", "method"}

	return &Prometheus{
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
		reqSize := computeApproximateRequestSize(r)

		start := time.Now()
		d := newDelegator(w, nil)
		h.ServeHTTP(d, r)
		elapsed := float64(time.Since(start)) / float64(time.Second)
		lv := [2]string{strconv.Itoa(d.Status()), r.Method}

		p.requestsTotal.WithLabelValues(lv[:]...).Inc()
		p.requestDuration.WithLabelValues(lv[:]...).Observe(elapsed)
		p.requestSize.WithLabelValues(lv[:]...).Observe(float64(reqSize))
		p.responseSize.WithLabelValues(lv[:]...).Observe(float64(d.Written()))
	})
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}

const durationKey = "sl-duration-key"

func (p *Prometheus) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	ctx.Set(durationKey, time.Now())
	return nil
}

func (p *Prometheus) ModifyResponse(res *http.Response) error {
	r := res.Request
	ctx := martian.NewContext(r)

	var elapsed float64
	if t0, ok := ctx.Get(durationKey); ok {
		start := t0.(time.Time) //nolint:forcetypeassert // we know it's time
		elapsed = float64(time.Since(start)) / float64(time.Second)
	} else {
		return fmt.Errorf("prometheus duration key not found")
	}

	reqSize := computeApproximateRequestSize(r)
	lv := [2]string{strconv.Itoa(res.StatusCode), r.Method}

	p.requestsTotal.WithLabelValues(lv[:]...).Inc()
	p.requestDuration.WithLabelValues(lv[:]...).Observe(elapsed)
	p.requestSize.WithLabelValues(lv[:]...).Observe(float64(reqSize))

	return nil
}
