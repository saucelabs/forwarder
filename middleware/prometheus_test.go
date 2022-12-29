// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
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

	golden, err := os.ReadFile("testdata/TestPrometheusWrap.golden.txt")
	if err != nil {
		t.Fatal(err)
	}

	got := dumpPrometheusMetrics(t, r)
	// Remove *_seconds_sum from the output, as it's not deterministic
	got = regexp.MustCompile(`(?m)^.*_seconds_sum.*$`).ReplaceAllString(got, "")

	if diff := cmp.Diff(string(golden), got); diff != "" {
		t.Errorf("unexpected metrics (-want +got):\n%s", diff)
		if err := os.WriteFile("testdata/TestPrometheusWrap.golden.txt", []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func dumpPrometheusMetrics(t *testing.T, r prometheus.Gatherer) string {
	t.Helper()

	got, err := r.Gather()
	if err != nil {
		t.Fatal(err)
	}
	var gotBuf bytes.Buffer
	enc := expfmt.NewEncoder(&gotBuf, expfmt.FmtText)
	for _, mf := range got {
		if err := enc.Encode(mf); err != nil {
			t.Fatal(err)
		}
	}
	return gotBuf.String()
}
