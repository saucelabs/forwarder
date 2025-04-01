// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

import (
	"context"
)

// NopLogger is a logger that does nothing.
var NopLogger = nopLogger{} //nolint:gochecknoglobals // nop implementation

var (
	_ Logger           = &nopLogger{}
	_ StructuredLogger = &nopLogger{}
)

type nopLogger struct{}

func (l nopLogger) Errorf(_ string, _ ...any) {}
func (l nopLogger) Infof(_ string, _ ...any)  {}
func (l nopLogger) Debugf(_ string, _ ...any) {}

func (l nopLogger) Error(_ string, _ ...any) {}
func (l nopLogger) Warn(_ string, _ ...any)  {}
func (l nopLogger) Info(_ string, _ ...any)  {}
func (l nopLogger) Debug(_ string, _ ...any) {}

func (l nopLogger) ErrorContext(_ context.Context, _ string, _ ...any) {}
func (l nopLogger) WarnContext(_ context.Context, _ string, _ ...any)  {}
func (l nopLogger) InfoContext(_ context.Context, _ string, _ ...any)  {}
func (l nopLogger) DebugContext(_ context.Context, _ string, _ ...any) {}

func (l nopLogger) With(_ ...any) StructuredLogger { return l }
