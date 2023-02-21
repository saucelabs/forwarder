// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/saucelabs/forwarder/utils/golden"
)

func TestPrometheusWrap(t *testing.T) {
	pages := []struct {
		path         string
		duration     time.Duration
		status       int
		requestSize  float64
		responseSize float64
	}{
		{"/1", 10 * time.Millisecond, http.StatusOK, 1.05 * kb, 1.05 * kb},
		{"/2", 100 * time.Millisecond, http.StatusOK, 5.05 * kb, 5.05 * kb},
		{"/3", 500 * time.Millisecond, http.StatusOK, 10.05 * kb, 10.05 * kb},
		{"/4", 1000 * time.Millisecond, http.StatusOK, 100.05 * kb, 100.05 * kb},
	}

	h := http.NewServeMux()
	for i := range pages {
		p := pages[i]
		h.HandleFunc(p.path, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(p.duration)
			w.WriteHeader(p.status)
			w.Write(make([]byte, int(p.responseSize)))
		})
	}

	r := prometheus.NewPedanticRegistry()
	s := NewPrometheus(r, "test").Wrap(h)

	for i := range pages {
		p := pages[i]
		r := httptest.NewRequest(http.MethodGet, p.path, bytes.NewBuffer(make([]byte, int(p.requestSize))))
		r.RemoteAddr = "localhost:1234"
		r.URL.Host = "saucelabs.com"
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
	}

	golden.DiffPrometheusMetrics(t, r, func(mf *dto.MetricFamily) bool {
		if int(*mf.Type) == 4 {
			for _, m := range mf.Metric {
				m.Histogram.SampleSum = nil
			}
		}
		return true
	})
}
