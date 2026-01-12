// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"math"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/saucelabs/forwarder/utils/golden"
)

func TestPrometheusWrap(t *testing.T) {
	pages := []struct {
		path     string
		duration time.Duration
	}{
		{"/100", 100 * time.Millisecond},
		{"/200", 200 * time.Millisecond},
		{"/1000", 1000 * time.Millisecond},
	}

	h := http.NewServeMux()
	for i := range pages {
		p := pages[i]
		h.HandleFunc(p.path, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(p.duration)
			w.WriteHeader(http.StatusOK)
		})
	}

	r := prometheus.NewPedanticRegistry()
	s := NewPrometheus(r, "test").Wrap(h)

	var wg sync.WaitGroup
	for range [100]struct{}{} {
		for i := range pages {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p := pages[i]
				r := httptest.NewRequest(http.MethodGet, p.path, http.NoBody)
				r.RemoteAddr = "localhost:1234"
				r.URL.Host = "saucelabs.com"
				w := httptest.NewRecorder()
				s.ServeHTTP(w, r)
			}()
		}
	}
	wg.Wait()

	oneDecimalDigit := func(v *float64) {
		*v = math.Round(*v*10) / 10
	}

	golden.DiffPrometheusMetrics(t, r, func(mf *dto.MetricFamily) bool {
		if int(mf.GetType()) == 2 {
			for _, m := range mf.GetMetric() {
				m.Summary.SampleSum = nil
				m.Summary.SampleCount = nil
				for _, q := range m.GetSummary().GetQuantile() {
					oneDecimalDigit(q.Value) //nolint:protogetter // We want to set the value.
				}
			}
		}
		return true
	})
}
