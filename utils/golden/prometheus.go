// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package golden

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// WaitMetrics is the time to wait for the metrics to be updated.
// Somehow, the metrics are not updated immediately at all times.
var WaitMetrics = 10 * time.Millisecond

func DiffPrometheusMetrics(t *testing.T, p prometheus.Gatherer, filter ...func(*dto.MetricFamily) bool) {
	t.Helper()

	time.Sleep(WaitMetrics)

	goldenFile := "testdata/" + strings.ReplaceAll(t.Name(), "/", "_") + ".golden.txt"
	golden, err := os.ReadFile(goldenFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		// Remove carriage returns from the file on Windows.
		golden = bytes.ReplaceAll(golden, []byte{'\r'}, nil)
	}

	got := dumpPrometheusMetrics(t, p, filter...)

	if diff := cmp.Diff(string(golden), got); diff != "" {
		t.Errorf("unexpected metrics (-want +got):\n%s", diff)
		if err := os.WriteFile(goldenFile, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func DiffPrometheusMetricsHTTP(t *testing.T, url string, filter ...func(*dto.MetricFamily) bool) {
	t.Helper()

	http.DefaultClient.Timeout = 30 * time.Second
	res, err := http.DefaultClient.Get(url) //nolint:noctx // The timeout is set above.
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	g, err := parseMetricFamilies(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	DiffPrometheusMetrics(t, g, filter...)
}

func dumpPrometheusMetrics(t *testing.T, p prometheus.Gatherer, filters ...func(*dto.MetricFamily) bool) string {
	t.Helper()

	got, err := p.Gather()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := range got {
		ok := true
		for _, f := range filters {
			if !f(mf) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}

		if err := enc.Encode(mf); err != nil {
			t.Fatal(err)
		}
	}
	return buf.String()
}

func parseMetricFamilies(reader io.Reader) (*gatherer, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	return &gatherer{mf: mf}, nil
}

type gatherer struct {
	mf map[string]*dto.MetricFamily
}

func (g *gatherer) Gather() ([]*dto.MetricFamily, error) {
	res := make([]*dto.MetricFamily, 0, len(g.mf))
	for _, mf := range g.mf {
		res = append(res, mf)
	}
	return res, nil
}
