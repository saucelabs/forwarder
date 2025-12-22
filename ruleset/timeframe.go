// Copyright 2022-2025 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ruleset

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TimeFrameEntry struct {
	Weekday   time.Weekday // Weekday is 0-6, going from Sunday to Saturday
	HourStart int          // 24h system
	HourEnd   int          // 24h system
}

func (t *TimeFrameEntry) Validate() error {
	if t.Weekday < 0 || t.Weekday > 6 {
		return errors.New("invalid weekday")
	}

	// We support right now only hour-based ranges, so for any computation
	// we assume minutes are set to 0. So technically 24:00:00 is a valid end hour
	// because there will be no way to set 23:59:59
	if t.HourStart < 0 || t.HourStart > 24 {
		return errors.New("HourStart outside valid range - <0,24>")
	}

	if t.HourEnd < 0 || t.HourEnd > 24 {
		return errors.New("HourEnd outside valid range - <0,24>")
	}

	if t.HourEnd < t.HourStart {
		return errors.New("HourEnd is earlier than HourStart")
	}

	return nil
}

// returns true if provided time matches TimeFrameEntry
// time is always converted to local time

func (t *TimeFrameEntry) Match(time time.Time) bool {

	// we expect time to already be in local time
	localTime := time

	if localTime.Weekday() != t.Weekday {
		return false
	}

	//  example:
	//  12-14 matches 12:00 until 13:59
	//  21-24 matches 21:00 until 23:59
	if localTime.Hour() >= t.HourStart && localTime.Hour() < t.HourEnd {
		return true
	}

	return false
}

func ParseTimeFrameEntry(repr string) (TimeFrameEntry, error) {
	// repr format: mon/11-13, wed/20-23, fri/23-24

	var newEntry TimeFrameEntry

	split := strings.Split(repr, "/")

	if len(split) != 2 || strings.TrimSpace(split[1]) == "" {
		return TimeFrameEntry{}, errors.New("invalid format")
	}

	weekdayName := strings.ToLower(strings.TrimSpace(split[0]))

	switch weekdayName {
	case "mon", "monday":
		newEntry.Weekday = time.Monday
	case "tue", "tuesday":
		newEntry.Weekday = time.Tuesday
	case "wed", "wednesday":
		newEntry.Weekday = time.Wednesday
	case "thu", "thursday":
		newEntry.Weekday = time.Thursday
	case "fri", "friday":
		newEntry.Weekday = time.Friday
	case "sat", "saturday":
		newEntry.Weekday = time.Saturday
	case "sun", "sunday":
		newEntry.Weekday = time.Sunday

	default:
		return TimeFrameEntry{}, errors.New("invalid weekday name")

	}

	hourStartHourEnd := strings.Split(split[1], "-")

	if len(hourStartHourEnd) != 2 {
		return TimeFrameEntry{}, errors.New("invalid format after '/'")
	}

	hourStart, err := strconv.Atoi(hourStartHourEnd[0])

	if err != nil {
		return TimeFrameEntry{}, fmt.Errorf("invalid format of start hour: %v", err)
	}

	newEntry.HourStart = hourStart

	hourEnd, err := strconv.Atoi(hourStartHourEnd[1])

	if err != nil {
		return TimeFrameEntry{}, fmt.Errorf("invalid format of end hour: %v", err)
	}

	newEntry.HourEnd = hourEnd

	// validate data at the end

	err = newEntry.Validate()

	if err != nil {
		return TimeFrameEntry{}, fmt.Errorf("time range entry value: %v", err)
	}

	return newEntry, nil
}
