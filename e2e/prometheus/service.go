// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package prometheus

import (
	"github.com/saucelabs/forwarder/utils/compose"
)

const (
	Image = "prom/prometheus:v2.47.2"
)

func Service() *compose.Service {
	return &compose.Service{
		Name:  "prom",
		Image: Image,
		Ports: []string{
			"9090:9090",
		},
		Volumes: []string{
			"./prometheus/prometheus.yaml:/etc/prometheus/prometheus.yml",
		},
	}
}
