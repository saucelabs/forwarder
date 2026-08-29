// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package slog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"

	flog "github.com/saucelabs/forwarder/log"
)

func Default() *Logger {
	return New(flog.DefaultConfig())
}

func Debug() *Logger {
	return New(&flog.Config{Level: flog.DebugLevel})
}

var _ flog.StructuredLogger = &Logger{}

type Option func(*Logger)

type Logger struct {
	log     *slog.Logger
	file    *flog.RotatableFile
	name    string
	onError func(name string)
}

func New(cfg *flog.Config, opts ...Option) *Logger {
	var w io.Writer = os.Stdout

	var f *flog.RotatableFile
	if cfg.File != nil {
		f = flog.NewRotatableFile(cfg.File)
		w = f
	}

	hops := &slog.HandlerOptions{
		Level: flogToSlogLevel(cfg.Level), AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == "source" {
				// 8 is a magic number to escape slog + forwarder internal call stack
				// this may change in the future depending on slog internals
				const callerDepth = 8
				pc := make([]uintptr, 1)
				retrievedCallers := runtime.Callers(callerDepth, pc)

				if retrievedCallers == 0 {
					// something wrong - most likely magic number changed and it's larger
					// than call stack depth. In this case log will probably show as source
					// logging call function somewhere in slog
					return attr
				}

				fs := runtime.CallersFrames([]uintptr{pc[0]})
				f, _ := fs.Next()

				return slog.Attr{
					Key:   attr.Key,
					Value: slog.StringValue(fmt.Sprintf("%s:%d", f.File, f.Line)),
				}
			}
			return attr
		},
	}
	var handler slog.Handler
	if cfg.Format == flog.JSONFormat {
		handler = slog.NewJSONHandler(w, hops)
	} else {
		handler = slog.NewTextHandler(w, hops)
	}
	logger := slog.New(handler)

	l := &Logger{
		log:  logger,
		file: f,
	}
	for _, opt := range opts {
		opt(l)
	}

	return l
}

func (l *Logger) Handler() slog.Handler {
	return l.log.Handler()
}

func (l *Logger) Error(msg string, args ...any) {
	if l.onError != nil {
		l.onError(l.name)
	}
	l.log.Error(msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	if l.onError != nil {
		l.onError(l.name)
	}
	l.log.ErrorContext(ctx, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log.WarnContext(ctx, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log.InfoContext(ctx, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log.DebugContext(ctx, msg, args...)
}

func (l *Logger) With(args ...any) flog.StructuredLogger {
	c := *l
	c.log = c.log.With(args...)
	return &c
}

func (l *Logger) Named(name string) *Logger {
	c := *l
	c.name = name
	c.log = c.log.With("module", name)
	return &c
}

func (l *Logger) Reopen() error {
	if l.file == nil {
		return nil
	}
	return l.file.Reopen()
}

func (l *Logger) Close() error {
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

func flogToSlogLevel(level flog.Level) slog.Level {
	switch level {
	case flog.ErrorLevel:
		return slog.LevelError
	case flog.WarnLevel:
		return slog.LevelWarn
	case flog.InfoLevel:
		return slog.LevelInfo
	case flog.DebugLevel:
		return slog.LevelDebug
	default:
		return slog.Level(level)
	}
}
