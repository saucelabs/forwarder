// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

// Logger is the logger used by the forwarder package.
type Logger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// NopLogger is a logger that does nothing.
var NopLogger = nopLogger{} //nolint:gochecknoglobals // nop implementation

type nopLogger struct{}

func (l nopLogger) Errorf(format string, args ...interface{}) {
}

func (l nopLogger) Infof(format string, args ...interface{}) {
}

func (l nopLogger) Debugf(format string, args ...interface{}) {
}
