// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package martianlog

import (
	"context"

	"github.com/saucelabs/forwarder/internal/martian"
	martianlog "github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/log"
)

func SetLogger(l log.Logger) {
	martianlog.SetLogger(martian.TraceIDPrependingLogger{
		Logger: contextLogger{l},
	})
}

// contextLogger is a wrapper around log.Logger that implements the martian log.Logger interface.
type contextLogger struct {
	log.Logger
}

var _ martianlog.Logger = contextLogger{}

func (l contextLogger) Errorf(_ context.Context, format string, args ...any) {
	l.Logger.Errorf(format, args...)
}

func (l contextLogger) Infof(_ context.Context, format string, args ...any) {
	l.Logger.Infof(format, args...)
}

func (l contextLogger) Debugf(_ context.Context, format string, args ...any) {
	l.Logger.Debugf(format, args...)
}
