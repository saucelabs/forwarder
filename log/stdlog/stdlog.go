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

func Default() Logger {
	return Logger{
		log:   log.Default(),
		level: flog.InfoLevel,
	}
}

func New(cfg *flog.Config) Logger {
	var w io.Writer = os.Stdout
	if cfg.File != nil {
		w = cfg.File
	}
	return Logger{
		log:   log.New(w, "", log.Ldate|log.Ltime|log.LUTC),
		level: cfg.Level,
	}
}

// Logger implements the forwarder.Logger interface using the standard log package.
type Logger struct {
	log   *log.Logger
	name  string
	level flog.Level

	// Decorate allows to modify the log message before it is written.
	Decorate func(string) string
}

func (sl Logger) Named(name string) Logger {
	if name != "" {
		name = "[" + name + "] "
	}
	sl.name = name
	return sl
}

func (sl Logger) Errorf(format string, args ...any) {
	if sl.level < flog.ErrorLevel {
		return
	}
	if sl.Decorate != nil {
		format = sl.Decorate(format)
	}
	sl.log.Printf(sl.name+"ERROR: "+format, args...)
}

func (sl Logger) Infof(format string, args ...any) {
	if sl.level < flog.InfoLevel {
		return
	}
	if sl.Decorate != nil {
		format = sl.Decorate(format)
	}
	sl.log.Printf(sl.name+"INFO: "+format, args...)
}

func (sl Logger) Debugf(format string, args ...any) {
	if sl.level < flog.DebugLevel {
		return
	}
	if sl.Decorate != nil {
		format = sl.Decorate(format)
	}
	sl.log.Printf(sl.name+"DEBUG: "+format, args...)
}

// Unwrap returns the underlying log.Logger pointer.
func (sl Logger) Unwrap() *log.Logger {
	return sl.log
}
