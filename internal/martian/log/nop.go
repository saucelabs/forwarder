// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.

package log

import (
	"context"
)

type nopLogger struct{}

var _ StructuredLogger = nopLogger{}

func (nopLogger) Error(_ context.Context, _ string, _ ...any) {}
func (nopLogger) Warn(_ context.Context, _ string, _ ...any)  {}
func (nopLogger) Info(_ context.Context, _ string, _ ...any)  {}
func (nopLogger) Debug(_ context.Context, _ string, _ ...any) {}

func (nopLogger) With(_ ...any) StructuredLogger { return nopLogger{} }
