// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

type Level int32

const (
	ErrorLevel Level = 1 + iota
	InfoLevel
	DebugLevel
)

func (l Level) String() string {
	return [3]string{"error", "info", "debug"}[l-1]
}
