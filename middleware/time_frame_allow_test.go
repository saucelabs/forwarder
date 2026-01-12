// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"testing"
	"time"

	"github.com/saucelabs/forwarder/ruleset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeFrameAllow(t *testing.T) {
	// comprehensive testing is done in TimeFrameEntry implementation
	// here we just check sanity of TimeFrameAllow

	var allowTimeFrame []ruleset.TimeFrameEntry
	strEntries := []string{"mon/10-13", "wed/22-23"}

	// January 1st 2025 was Wednesday - this will be important later

	mockedTime := time.Date(2025, time.January, 1, 22, 10, 0, 0, time.UTC)

	mockedGetCurrentTime := func() time.Time {
		return mockedTime
	}

	getCurrentTime = mockedGetCurrentTime

	for _, strEntry := range strEntries {
		parsedEntry, err := ruleset.ParseTimeFrameEntry(strEntry)
		require.NoError(t, err)
		allowTimeFrame = append(allowTimeFrame, parsedEntry)
	}

	assert.True(t, TimeFrameAllows(allowTimeFrame))

	// time within first rule - Jan 6 is Monday
	mockedTime = time.Date(2025, time.January, 6, 10, 15, 10, 0, time.UTC)
	assert.True(t, TimeFrameAllows(allowTimeFrame))

	// time outside allowed time frame
	mockedTime = time.Date(2025, time.January, 1, 19, 10, 0, 0, time.UTC)
	assert.False(t, TimeFrameAllows(allowTimeFrame))

	getCurrentTime = time.Now
}
