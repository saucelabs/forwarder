// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dns

import (
	"github.com/saucelabs/forwarder/utils/compose"
)

const (
	ServiceName = "dns"
	Image       = "e2e-dns"
)

func Service(network, ipv4 string) *compose.Service {
	return &compose.Service{
		Name:  ServiceName,
		Image: Image,
		Network: map[string]compose.ServiceNetwork{
			network: {IPv4: ipv4},
		},
	}
}
