// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package martianlog

import (
	"context"
	"fmt"
	"os"

	"github.com/saucelabs/forwarder/log"
)

func newStructuredLoggerAdapter(log log.Logger) *structuredLoggerAdapter {
	return &structuredLoggerAdapter{log: log}
}

type structuredLoggerAdapter struct {
	log log.Logger
}

func (la *structuredLoggerAdapter) FatalContext(_ context.Context, msg string, args ...any) {
	la.log.Errorf("[FATAL] %s", formatMessage(msg, args...))
	os.Exit(1)
}

func (la *structuredLoggerAdapter) ErrorContext(_ context.Context, msg string, args ...any) {
	la.log.Errorf("%s", formatMessage(msg, args...))
}

func (la *structuredLoggerAdapter) WarnContext(_ context.Context, msg string, args ...any) {
	la.log.Errorf("[WARN] %s", formatMessage(msg, args...))
}

func (la *structuredLoggerAdapter) InfoContext(_ context.Context, msg string, args ...any) {
	la.log.Infof("%s", formatMessage(msg, args...))
}

func (la *structuredLoggerAdapter) DebugContext(_ context.Context, msg string, args ...any) {
	la.log.Debugf("%s", formatMessage(msg, args...))
}

func (la *structuredLoggerAdapter) TraceContext(_ context.Context, msg string, args ...any) {
	la.log.Debugf("[TRACE] %s", formatMessage(msg, args...))
}

// formatMessage converts a message and its arguments into a single string formatted as:
// "msg arg0=arg1 arg2=arg3 ...".
func formatMessage(msg string, args ...any) string {
	result := msg
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			result += " " + fmt.Sprintf("%v=%v", args[i], args[i+1])
		} else {
			result += " " + fmt.Sprintf("%v", args[i])
		}
	}
	return result
}
