// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

import (
	"os"
)

// Config is a configuration for the loggers.
type Config struct {
	File    *os.File
	Verbose bool
}

func DefaultConfig() *Config {
	return &Config{
		File: nil,
	}
}
