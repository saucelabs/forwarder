// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.

package log

import (
	"context"
)

type StructuredLogger interface {
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)

	With(args ...any) StructuredLogger
}

var log StructuredLogger = nopLogger{}

// SetLogger changes the default logger. This must be called very first,
// before interacting with rest of the martian package. Changing it at
// runtime is not supported.
func SetLogger(l StructuredLogger) {
	log = l
}

func Error(ctx context.Context, msg string, args ...any) {
	log.Error(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	log.Warn(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	log.Info(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	log.Debug(ctx, msg, args...)
}
