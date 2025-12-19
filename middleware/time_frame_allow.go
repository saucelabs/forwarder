// Copyright 2022-2025 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"time"

	"github.com/saucelabs/forwarder/ruleset"
)

var getCurrentTime = time.Now

// returns true if current time matches any entry in a list of allowed
// timeFrameEntry objects
func TimeFrameAllows(allowRules []ruleset.TimeFrameEntry) bool {

	currentTime := getCurrentTime()

	for _, rule := range allowRules {
		if rule.Match(currentTime) {
			return true
		}
	}
	return false

}
