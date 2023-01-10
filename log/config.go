// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

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
