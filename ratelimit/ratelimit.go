// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ratelimit

import (
	"golang.org/x/time/rate"
)

const defaultMaxBurstSize = 4 * 1024 * 1024 // Must be bigger than the biggest request.

func newRateLimiter(bandwidth int64) *rate.Limiter {
	// Relate maxBurstSize to bandwidth limit
	// 4M gives 2.5 Gb/s on Windows
	// Use defaultMaxBurstSize up to 2GBit/s (256MiB/s) then scale
	// https://github.com/rclone/rclone/issues/5507
	maxBurstSize := bandwidth / 64
	if maxBurstSize < defaultMaxBurstSize {
		maxBurstSize = defaultMaxBurstSize
	}
	return rate.NewLimiter(rate.Limit(bandwidth), int(maxBurstSize))
}
