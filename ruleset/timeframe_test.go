// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ruleset

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeFrameInvalidFormat(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{
			input:         "abc123",
			expectedError: "invalid format",
		},
		{
			input:         "foo/",
			expectedError: "invalid format",
		},
		{
			input:         "////",
			expectedError: "invalid format",
		},
		{
			input:         "xyz/12-13",
			expectedError: "invalid weekday name",
		},
		{
			input:         "mon/80-85",
			expectedError: "time range entry value: HourStart outside valid range - <0,24>",
		},
		{
			input:         "mon/13-80",
			expectedError: "time range entry value: HourEnd outside valid range - <0,24>",
		},
		{
			input:         "mon/12-10",
			expectedError: "time range entry value: HourEnd is earlier than HourStart",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.input, func(t *testing.T) {
			_, err := ParseTimeFrameEntry(tc.input)
			require.Error(t, err)
			assert.Equal(t, tc.expectedError, err.Error())
		})
	}
}

func TestTimeParseWeekday(t *testing.T) {
	tests := []struct {
		input           string
		expectedWeekDay time.Weekday
	}{
		{
			input:           "mon/11-13",
			expectedWeekDay: time.Monday,
		},
		{
			input:           "MoN/11-11",
			expectedWeekDay: time.Monday,
		},
		{
			input:           "MoNdAy/11-13",
			expectedWeekDay: time.Monday,
		},
		{
			input:           "tue/11-13",
			expectedWeekDay: time.Tuesday,
		},
		{
			input:           "tuesday/11-13",
			expectedWeekDay: time.Tuesday,
		},
		{
			input:           "wed/11-13",
			expectedWeekDay: time.Wednesday,
		},
		{
			input:           "wednesday/11-13",
			expectedWeekDay: time.Wednesday,
		},
		{
			input:           "thu/11-13",
			expectedWeekDay: time.Thursday,
		},
		{
			input:           "thursday/11-13",
			expectedWeekDay: time.Thursday,
		},
		{
			input:           "fri/11-13",
			expectedWeekDay: time.Friday,
		},
		{
			input:           "sat/11-13",
			expectedWeekDay: time.Saturday,
		},
		{
			input:           "saturday/11-13",
			expectedWeekDay: time.Saturday,
		},
		{
			input:           "sun/11-13",
			expectedWeekDay: time.Sunday,
		},
		{
			input:           "sunday/11-13",
			expectedWeekDay: time.Sunday,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.input, func(t *testing.T) {
			timeframe, err := ParseTimeFrameEntry(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedWeekDay, timeframe.Weekday)
		})
	}
}

func TestTimeParseHours(t *testing.T) {
	tests := []struct {
		input             string
		expectedHourStart int
		expectedHourEnd   int
	}{
		{
			input:             "mon/11-13",
			expectedHourStart: 11,
			expectedHourEnd:   13,
		},
		{
			input:             "mon/0-24",
			expectedHourStart: 0,
			expectedHourEnd:   24,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.input, func(t *testing.T) {
			timeframe, err := ParseTimeFrameEntry(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedHourStart, timeframe.HourStart)
			assert.Equal(t, tc.expectedHourEnd, timeframe.HourEnd)
		})
	}
}

func TestTimeFrameMatch(t *testing.T) {
	// January 1st 2025 was Wednesday - this will be important later

	tests := []struct {
		currentTime time.Time
		input       string
		shouldMatch bool
	}{
		{
			currentTime: time.Date(2025, time.January, 1, 10, 10, 0, 0, time.UTC),
			input:       "wed/10-13",
			shouldMatch: true,
		},
		{
			currentTime: time.Date(2025, time.January, 2, 10, 10, 0, 0, time.UTC),
			input:       "wed/10-13",
			shouldMatch: false, // wrong day of week 2nd is Thursday
		},
		{
			// testing hourStart exact match (10-13 is 10:00:00-12:59:59)
			currentTime: time.Date(2025, time.January, 2, 10, 0, 0, 0, time.UTC),
			input:       "thu/10-13",
			shouldMatch: true,
		},
		{
			// testing hourEnd non-including range (ie. it ends 12:59)
			currentTime: time.Date(2025, time.January, 2, 13, 0, 0, 0, time.UTC),
			input:       "thu/10-13",
			shouldMatch: false,
		},
		{
			// testing hourEnd non-including range (ie. it ends 12:59)
			currentTime: time.Date(2025, time.January, 2, 12, 59, 59, 0, time.UTC),
			input:       "thu/10-13",
			shouldMatch: true,
		},
		{
			// 0-12 left range should match 00:00:00:00
			currentTime: time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC),
			input:       "thu/0-12",
			shouldMatch: true,
		},
		{
			// 12-24 right range should not match 00:00:00:00
			currentTime: time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC),
			input:       "thu/12-24",
			shouldMatch: false,
		},
		{
			// 12-24 right range should match 23:59:59:00
			currentTime: time.Date(2025, time.January, 2, 23, 59, 59, 0, time.UTC),
			input:       "thu/12-24",
			shouldMatch: true,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.input, func(t *testing.T) {
			timeframe, err := ParseTimeFrameEntry(tc.input)
			require.NoError(t, err)
			assert.Equal(t, timeframe.Match(tc.currentTime), tc.shouldMatch)
		})
	}
}
