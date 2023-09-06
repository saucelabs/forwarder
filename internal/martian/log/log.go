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

// Package log provides a universal logger for martian packages.
package log

type Logger interface {
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

var currLogger Logger = nopLogger{}

// SetLogger changes the default logger. This must be called very first,
// before interacting with rest of the martian package. Changing it at
// runtime is not supported.
func SetLogger(l Logger) {
	currLogger = l
}

// Infof logs an info message.
func Infof(format string, args ...any) {
	currLogger.Infof(format, args...)
}

// Debugf logs a debug message.
func Debugf(format string, args ...any) {
	currLogger.Debugf(format, args...)
}

// Errorf logs an error message.
func Errorf(format string, args ...any) {
	currLogger.Errorf(format, args...)
}

type nopLogger struct{}

func (nopLogger) Infof(_ string, _ ...any) {}

func (nopLogger) Debugf(_ string, _ ...any) {}

func (nopLogger) Errorf(_ string, _ ...any) {}
