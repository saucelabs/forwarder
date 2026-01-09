// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"
	"log/slog"
)

var _ StructuredLogger = testLogger{}

type testLogger struct {
	*slog.Logger
}

func (l testLogger) Error(ctx context.Context, msg string, args ...any) {
	l.ErrorContext(ctx, msg, args...)
}

func (l testLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.WarnContext(ctx, msg, args...)
}

func (l testLogger) Info(ctx context.Context, msg string, args ...any) {
	l.InfoContext(ctx, msg, args...)
}

func (l testLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.DebugContext(ctx, msg, args...)
}

func (l testLogger) With(args ...any) StructuredLogger {
	return testLogger{l.Logger.With(args...)}
}

func SetTestLogger() {
	SetLogger(testLogger{slog.Default()})
}
