// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	_          = iota // ignore first value by assigning to blank identifier
	kb float64 = 1 << (10 * iota)
	mb
)

func TestPrometheusWrap(t *testing.T) {
	pages := []struct {
		path     string
		duration time.Duration
		status   int
		size     float64
	}{
		{"/1", 10 * time.Millisecond, http.StatusOK, 1.05 * kb},
		{"/2", 100 * time.Millisecond, http.StatusOK, 5.05 * kb},
		{"/3", 500 * time.Millisecond, http.StatusOK, 10.05 * kb},
		{"/4", 1000 * time.Millisecond, http.StatusOK, 100.05 * kb},
	}

	h := http.NewServeMux()
	for i := range pages {
		p := pages[i]
		h.HandleFunc(p.path, func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(p.duration)
			w.WriteHeader(p.status)
			n, err := io.Copy(w, r.Body)
			if err != nil {
				t.Error(err)
			}
			if n != int64(p.size) {
				t.Errorf("expected %d, got %d", int(p.size), n)
			}
		})
	}

	r := prometheus.NewPedanticRegistry()
	s := NewPrometheus(r, "test").Wrap(h)

	for i := range pages {
		p := pages[i]
		b := bytes.NewBuffer(make([]byte, int(p.size)))
		r := httptest.NewRequest(http.MethodGet, p.path, b)
		r.RemoteAddr = "localhost:1234"
		r.URL.Host = "saucelabs.com"
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
	}
}
