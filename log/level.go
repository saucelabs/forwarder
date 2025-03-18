// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

type Level int

// Levels start from 1 to avoid zero value in help printer.
const (
	TraceLevel Level = 1 + iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l Level) String() string {
	return [6]string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[l-1]
}
