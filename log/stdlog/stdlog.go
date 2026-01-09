// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package stdlog

import (
	"io"
	"log"
	"os"
	"strings"

	flog "github.com/saucelabs/forwarder/log"
)

func Default() *Logger {
	return &Logger{
		log:   log.Default(),
		level: flog.InfoLevel,
	}
}

func Debug() *Logger {
	return &Logger{
		log:   log.Default(),
		level: flog.DebugLevel,
	}
}

// Option is a function that modifies the Logger.
type Option func(*Logger)

func New(cfg *flog.Config, opts ...Option) *Logger {
	var w io.Writer = os.Stdout

	var f *flog.RotatableFile
	if cfg.File != nil {
		f = flog.NewRotatableFile(cfg.File)
		w = f
	}

	l := Logger{
		log:   log.New(w, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC),
		file:  f,
		level: cfg.Level,
	}
	for _, opt := range opts {
		opt(&l)
	}
	return l.Named("")
}

// Logger implements the forwarder.Logger interface using the standard log package.
//
//nolint:recvcheck // This is intentional.
type Logger struct {
	log    *log.Logger
	file   *flog.RotatableFile
	labels []string
	name   string
	level  flog.Level

	errorPfx string
	infoPfx  string
	debugPfx string

	decorate func(string) string
	onError  func(name string)
}

func (sl Logger) Named(name string, opts ...Option) *Logger { //nolint:gocritic // we pass by value to get a copy
	sl.name = name

	sl.errorPfx = logLinePrefix(sl.labels, name, "ERROR")
	sl.infoPfx = logLinePrefix(sl.labels, name, "INFO")
	sl.debugPfx = logLinePrefix(sl.labels, name, "DEBUG")

	for _, opt := range opts {
		opt(&sl)
	}

	return &sl
}

func logLinePrefix(labels []string, name, level string) string {
	all := append(labels[0:len(labels):len(labels)], name, level) //nolint:gocritic // all is good we always create new slice
	var sb strings.Builder
	for _, l := range all {
		if l == "" {
			continue
		}
		sb.WriteString("[")
		sb.WriteString(l)
		sb.WriteString("] ")
	}
	return sb.String()
}

func (sl *Logger) Errorf(format string, args ...any) {
	if sl.onError != nil {
		defer sl.onError(sl.name)
	}
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

func (sl *Logger) Reopen() error {
	if sl.file == nil {
		return nil
	}
	return sl.file.Reopen()
}

func (sl *Logger) Close() error {
	if sl.file == nil {
		return nil
	}
	return sl.file.Close()
}
