// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package promutil

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type Desc struct {
	FqName         string
	Help           string
	ConstLabels    map[string]string
	VariableLabels []string
}

func DescribePrometheusMetrics(p prometheus.Collector) []Desc {
	ch := make(chan *prometheus.Desc, 1)
	go func() {
		p.Describe(ch)
		close(ch)
	}()

	var res []Desc //nolint:prealloc // We don't know the size of the result
	for d := range ch {
		res = append(res, parseDesc(d.String()))
	}
	return res
}

func parseDesc(s string) Desc {
	var res Desc
	res.FqName, res.Help = parseDescNameAndHelp(s)
	res.ConstLabels = parseDescConstLabels(s)
	res.VariableLabels = parseDescVariableLabels(s)
	return res
}

func parseDescNameAndHelp(s string) (name, help string) {
	fmt.Fscanf(strings.NewReader(s), "Desc{fqName: %q, help: %q}", &name, &help) //nolint:errcheck // reading from a string can't fail
	return
}

func parseDescConstLabels(s string) map[string]string {
	const pfx = "constLabels: {"
	const sfx = "}"

	start := strings.Index(s, pfx) + len(pfx)
	end := start + strings.Index(s[start:], sfx)
	s = s[start:end]

	if s == "" {
		return nil
	}

	res := make(map[string]string)
	for _, kv := range strings.Split(s, ",") {
		k, v, _ := strings.Cut(kv, "=")
		if k != "" {
			res[k] = strings.Trim(v, "\"")
		}
	}

	return res
}

func parseDescVariableLabels(s string) []string {
	const pfx = "variableLabels: {"
	const sfx = "}"

	start := strings.Index(s, pfx) + len(pfx)
	end := start + strings.Index(s[start:], sfx)
	s = s[start:end]

	if s == "" {
		return nil
	}

	return strings.Split(s, ",")
}
