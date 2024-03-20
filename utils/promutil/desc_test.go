// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package promutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func TestDescribePrometheusMetrics(t *testing.T) {
	r := prometheus.NewRegistry()
	f := promauto.With(r)

	f.NewGauge(prometheus.GaugeOpts{ //nolint:promlinter // For test purposes
		Name: "test_gauge",
		Help: "Test gauge",
		ConstLabels: prometheus.Labels{
			"const1": "label",
			"const2": "label",
		},
	})
	f.NewGaugeVec(prometheus.GaugeOpts{ //nolint:promlinter // For test purposes
		Name: "test_gauge_vec",
		Help: "Test gauge vec",
	}, []string{"label1", "label2"})
	f.NewGaugeFunc(prometheus.GaugeOpts{ //nolint:promlinter // For test purposes
		Name: "test_gauge_func",
		Help: "Test gauge func",
	}, func() float64 {
		return 0
	})
	f.NewCounter(prometheus.CounterOpts{ //nolint:promlinter // For test purposes
		Name: "test_counter",
		Help: "Test counter",
	})
	f.NewCounterVec(prometheus.CounterOpts{ //nolint:promlinter // For test purposes
		Name: "test_counter_vec",
		Help: "Test counter vec",
	}, []string{"label"})

	golden := []Desc{
		{
			FqName: "test_gauge",
			Help:   "Test gauge",
			ConstLabels: map[string]string{
				"const1": "label",
				"const2": "label",
			},
		},
		{
			FqName:         "test_counter_vec",
			Help:           "Test counter vec",
			VariableLabels: []string{"label"},
		},
		{
			FqName: "test_gauge_func",
			Help:   "Test gauge func",
		},
		{
			FqName:         "test_gauge_vec",
			Help:           "Test gauge vec",
			VariableLabels: []string{"label1", "label2"},
		},
		{
			FqName: "test_counter",
			Help:   "Test counter",
		},
	}

	desc := DescribePrometheusMetrics(r)

	sf := func(a, b Desc) bool {
		return a.FqName < b.FqName
	}
	if diff := cmp.Diff(golden, desc, cmpopts.SortSlices(sf)); diff != "" {
		t.Errorf("unexpected metrics (-want +got):\n%s", diff)
	}
}
