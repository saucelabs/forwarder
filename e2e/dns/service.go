// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dns

import (
	"github.com/saucelabs/forwarder/utils/compose"
)

type service compose.Service

const (
	Image = "ubuntu/bind9:latest"

	ServiceName = "dns"
)

func Service() *service { //nolint:revive,golint // Unexported by design.
	return &service{
		Name:        ServiceName,
		Image:       Image,
		Environment: map[string]string{},
		Volumes: []string{
			"./dns/db.rpz:/etc/bind/db.rpz",
			"./dns/named.conf.local:/etc/bind/named.conf.local",
			"./dns/named.conf.options:/etc/bind/named.conf.options",
		},
	}
}

func (s *service) WithIP(network, ipv4 string) *service {
	s.Network = map[string]compose.ServiceNetwork{
		network: {IPv4: ipv4},
	}
	return s
}

func (s *service) Service() *compose.Service {
	return (*compose.Service)(s)
}
