// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package promutil

import (
	"bytes"
	"io"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func DumpPrometheusMetrics(p prometheus.Gatherer, filters ...func(*dto.MetricFamily) bool) (string, error) {
	got, err := p.Gather()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
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
			return "", err
		}
	}

	return buf.String(), nil
}

func ParseMetricFamilies(reader io.Reader) (*Gatherer, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	return &Gatherer{mf: mf}, nil
}

type Gatherer struct {
	mf map[string]*dto.MetricFamily
}

func (g *Gatherer) Gather() ([]*dto.MetricFamily, error) {
	res := make([]*dto.MetricFamily, 0, len(g.mf))
	for _, mf := range g.mf {
		res = append(res, mf)
	}
	return res, nil
}
