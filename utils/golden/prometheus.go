// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package golden

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func DiffPrometheusMetrics(t *testing.T, p prometheus.Gatherer, filter ...func(*dto.MetricFamily) bool) {
	t.Helper()

	goldenFile := "testdata/" + t.Name() + ".golden.txt"
	golden, err := os.ReadFile(goldenFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}

	got := dumpPrometheusMetrics(t, p, filter...)

	if diff := cmp.Diff(string(golden), got); diff != "" {
		t.Errorf("unexpected metrics (-want +got):\n%s", diff)
		if err := os.WriteFile(goldenFile, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
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
