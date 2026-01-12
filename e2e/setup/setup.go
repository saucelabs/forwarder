// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setup

import (
	"github.com/saucelabs/forwarder/utils/compose"
)

const TestServiceName = "test"

type Setup struct {
	Name    string
	Compose *compose.Compose
	Run     string
}
