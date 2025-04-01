// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// NewLoggerAdapter returns a new StructuredLogger adapter for the given old Logger.
func NewLoggerAdapter(log Logger) StructuredLogger {
	return &loggerAdapter{log: log}
}

// loggerAdapter wraps an old Logger and implements the StructuredLogger interface.
type loggerAdapter struct {
	log    Logger
	fields []string
}

func (la *loggerAdapter) Fatal(msg string, args ...any) {
	formatted := la.buildMessage("FATAL: "+msg, args...)
	la.log.Errorf("%s", formatted)
	os.Exit(1)
}

func (la *loggerAdapter) Error(msg string, args ...any) {
	formatted := la.buildMessage(msg, args...)
	la.log.Errorf("%s", formatted)
}

func (la *loggerAdapter) Warn(msg string, args ...any) {
	formatted := la.buildMessage("WARN: "+msg, args...)
	la.log.Infof("%s", formatted)
}

func (la *loggerAdapter) Info(msg string, args ...any) {
	formatted := la.buildMessage(msg, args...)
	la.log.Infof("%s", formatted)
}

func (la *loggerAdapter) Debug(msg string, args ...any) {
	formatted := la.buildMessage(msg, args...)
	la.log.Debugf("%s", formatted)
}

func (la *loggerAdapter) Trace(msg string, args ...any) {
	formatted := la.buildMessage("TRACE: "+msg, args...)
	la.log.Debugf("%s", formatted)
}

func (la *loggerAdapter) FatalContext(ctx context.Context, msg string, args ...any) {
	la.Fatal(msg, args...)
}

func (la *loggerAdapter) ErrorContext(ctx context.Context, msg string, args ...any) {
	la.Error(msg, args...)
}

func (la *loggerAdapter) WarnContext(ctx context.Context, msg string, args ...any) {
	la.Warn(msg, args...)
}

func (la *loggerAdapter) InfoContext(ctx context.Context, msg string, args ...any) {
	la.Info(msg, args...)
}

func (la *loggerAdapter) DebugContext(ctx context.Context, msg string, args ...any) {
	la.Debug(msg, args...)
}

func (la *loggerAdapter) TraceContext(ctx context.Context, msg string, args ...any) {
	la.Trace(msg, args...)
}

// buildMessage constructs the final log message by combining the base message,
// call-specific fields (formatted from args), and the stored fields.
func (la *loggerAdapter) buildMessage(msg string, args ...any) string {
	callFields := formatMessage("", args...)
	storedFields := ""
	if len(la.fields) > 0 {
		storedFields = " " + strings.Join(la.fields, " ")
	}
	return msg + callFields + storedFields
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

func (la *loggerAdapter) With(args ...any) StructuredLogger {
	newField := formatFields(args...)
	newFields := make([]string, len(la.fields))
	copy(newFields, la.fields)
	if newField != "" {
		newFields = append(newFields, newField)
	}
	return &loggerAdapter{
		log:    la.log,
		fields: newFields,
	}
}

func formatFields(args ...any) string {
	s := formatMessage("", args...)
	if strings.HasPrefix(s, " ") {
		return s[1:]
	}
	return s
}
