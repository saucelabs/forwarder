// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

import (
	"context"
	"os"
)

// Logger is the logger used by the forwarder package.
type Logger interface {
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
}

// StructuredLogger is the preferred logging interface.
type StructuredLogger interface {
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)

	ErrorContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)

	With(args ...any) StructuredLogger
}

// NopLogger is a logger that does nothing.
var NopLogger = nopLogger{} //nolint:gochecknoglobals // nop implementation

type nopLogger struct{}

func (l nopLogger) Errorf(_ string, _ ...any) {
}

func (l nopLogger) Infof(_ string, _ ...any) {
}

func (l nopLogger) Debugf(_ string, _ ...any) {
}

var (
	DefaultFileFlags = os.O_CREATE | os.O_APPEND | os.O_WRONLY

	DefaultFileMode os.FileMode = 0o600
	DefaultDirMode  os.FileMode = 0o700
)
