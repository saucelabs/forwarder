// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.

package log

import (
	"context"
)

type Logger interface {
	Infof(ctx context.Context, format string, args ...any)
	Debugf(ctx context.Context, format string, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
}

var log Logger = nopLogger{}

// SetLogger changes the default logger. This must be called very first,
// before interacting with rest of the martian package. Changing it at
// runtime is not supported.
func SetLogger(l Logger) {
	log = l
}

// Infof logs an info message.
func Infof(ctx context.Context, format string, args ...any) {
	log.Infof(ctx, format, args...)
}

// Debugf logs a debug message.
func Debugf(ctx context.Context, format string, args ...any) {
	log.Debugf(ctx, format, args...)
}

// Errorf logs an error message.
func Errorf(ctx context.Context, format string, args ...any) {
	log.Errorf(ctx, format, args...)
}
