// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.

package log

import (
	"context"
)

type nopLogger struct{}

var _ StructuredLogger = nopLogger{}

func (nopLogger) ErrorContext(_ context.Context, _ string, _ ...any) {}
func (nopLogger) WarnContext(_ context.Context, _ string, _ ...any)  {}
func (nopLogger) InfoContext(_ context.Context, _ string, _ ...any)  {}
func (nopLogger) DebugContext(_ context.Context, _ string, _ ...any) {}

func (nopLogger) With(_ ...any) StructuredLogger { return nopLogger{} }
