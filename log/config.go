// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
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
	File   *os.File
	Level  Level
	Format Format
}

func DefaultConfig() *Config {
	return &Config{
		File:   nil,
		Level:  InfoLevel,
		Format: TextFormat,
	}
}

type Level int

// Levels start from 1 to avoid zero value in help printer.
const (
	ErrorLevel Level = 1 + iota
	WarnLevel
	InfoLevel
	DebugLevel
)

func (l Level) String() string {
	return [4]string{"error", "warn", "info", "debug"}[l-1]
}

type Format int

// Formats start from 1 to avoid zero value in help printer.
const (
	TextFormat Format = 1 + iota
	JSONFormat
)

func (m Format) String() string {
	return [2]string{"text", "json"}[m-1]
}
