// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.

package log

import (
	"context"
)

type StructuredLogger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)

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
	log.ErrorContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	log.WarnContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	log.InfoContext(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	log.DebugContext(ctx, msg, args...)
}
