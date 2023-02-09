// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

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
