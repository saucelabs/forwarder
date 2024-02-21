// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"github.com/prometheus/client_golang/prometheus"
)

// promConfig is a configuration for Prometheus metrics.
type promConfig struct {
	PromNamespace string
	PromRegistry  prometheus.Registerer
}
