// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

type Level string

const (
	ErrorLevel Level = "error"
	InfoLevel  Level = "info"
	DebugLevel Level = "debug"
)

func (l Level) String() string {
	return string(l)
}
