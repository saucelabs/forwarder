// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build dnshack

package forwarder

import (
	"github.com/saucelabs/forwarder/utils/dnshack"
)

func (c *DNSConfig) Apply() error {
	return dnshack.Configure(c.Servers, c.Timeout, c.RoundRobin)
}
