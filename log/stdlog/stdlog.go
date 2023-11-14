// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package stdlog

import (
	"io"
	"log"
	"os"

	flog "github.com/saucelabs/forwarder/log"
)

func Default() *Logger {
	return &Logger{
		log:   log.Default(),
		level: flog.InfoLevel,
	}
}

// Option is a function that modifies the Logger.
type Option func(*Logger)

func New(cfg *flog.Config, opts ...Option) *Logger {
	var w io.Writer = os.Stdout
	if cfg.File != nil {
		w = cfg.File
	}

	l := &Logger{
		log:   log.New(w, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC),
		level: cfg.Level,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Logger implements the forwarder.Logger interface using the standard log package.
type Logger struct {
	log   *log.Logger
	name  string
	level flog.Level

	errorPfx string
	infoPfx  string
	debugPfx string

	decorate func(string) string
}

func (sl Logger) Named(name string) *Logger { //nolint:gocritic // we pass by value to get a copy
	sl.name = name

	if name != "" {
		name = "[" + name + "] "
	}

	sl.errorPfx = name + "[ERROR] "
	sl.infoPfx = name + "[INFO] "
	sl.debugPfx = name + "[DEBUG] "

	return &sl
}

func (sl *Logger) Errorf(format string, args ...any) {
	if sl.level < flog.ErrorLevel {
		return
	}
	if sl.decorate != nil {
		format = sl.decorate(format)
	}
	sl.log.Printf(sl.errorPfx+format, args...)
}

func (sl *Logger) Infof(format string, args ...any) {
	if sl.level < flog.InfoLevel {
		return
	}
	if sl.decorate != nil {
		format = sl.decorate(format)
	}
	sl.log.Printf(sl.infoPfx+format, args...)
}

func (sl *Logger) Debugf(format string, args ...any) {
	if sl.level < flog.DebugLevel {
		return
	}
	if sl.decorate != nil {
		format = sl.decorate(format)
	}
	sl.log.Printf(sl.debugPfx+format, args...)
}

// Unwrap returns the underlying log.Logger pointer.
func (sl *Logger) Unwrap() *log.Logger {
	return sl.log
}
