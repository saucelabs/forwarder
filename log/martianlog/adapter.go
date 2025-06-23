// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package martianlog

import (
	"context"
	"fmt"

	martianlog "github.com/saucelabs/forwarder/internal/martian/log"
	"github.com/saucelabs/forwarder/log"
)

func newStructuredLoggerAdapter(log log.Logger) *structuredLoggerAdapter {
	return &structuredLoggerAdapter{log: log}
}

type structuredLoggerAdapter struct {
	log  log.Logger
	args []any
}

func (l *structuredLoggerAdapter) Error(_ context.Context, msg string, args ...any) {
	l.log.Errorf("%s", formatMessage(msg, append(l.args, args...)...))
}

func (l *structuredLoggerAdapter) Warn(_ context.Context, msg string, args ...any) {
	l.log.Infof("[WARN] %s", formatMessage(msg, append(l.args, args...)...))
}

func (l *structuredLoggerAdapter) Info(_ context.Context, msg string, args ...any) {
	l.log.Infof("%s", formatMessage(msg, append(l.args, args...)...))
}

func (l *structuredLoggerAdapter) Debug(_ context.Context, msg string, args ...any) {
	l.log.Debugf("%s", formatMessage(msg, append(l.args, args...)...))
}

func (l *structuredLoggerAdapter) With(args ...any) martianlog.StructuredLogger {
	return &structuredLoggerAdapter{
		log:  l.log,
		args: append(l.args, args...),
	}
}

// formatMessage converts a message and its arguments into a single string formatted as:
// "msg arg0=arg1 arg2=arg3 ...".
func formatMessage(msg string, args ...any) string {
	result := msg
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if args[i] == "id" { // Use old id format.
				result = "[" + fmt.Sprintf("%v", args[i+1]) + "] " + result
			} else {
				result += " " + fmt.Sprintf("%v=%v", args[i], args[i+1])
			}
		} else {
			result += " " + fmt.Sprintf("%v", args[i])
		}
	}
	return result
}
