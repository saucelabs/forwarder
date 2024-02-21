// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

// Logger is the logger used by the forwarder package.
type Logger interface {
	Errorf(format string, args ...any)
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
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
