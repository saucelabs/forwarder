// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package martianlog

import (
	martianlog "github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/log"
)

func newStructuredLoggerAdapter(log log.StructuredLogger) *structuredLoggerAdapter {
	return &structuredLoggerAdapter{log}
}

type structuredLoggerAdapter struct {
	log.StructuredLogger
}

func (l *structuredLoggerAdapter) With(args ...any) martianlog.StructuredLogger {
	slog := l.StructuredLogger.With(args...)
	return &structuredLoggerAdapter{slog}
}
