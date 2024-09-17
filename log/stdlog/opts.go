// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package stdlog

import flog "github.com/saucelabs/forwarder/log"

// WithLabels allows to set labels that are added to each log message.
func WithLabels(labels ...string) Option {
	return func(l *Logger) {
		l.labels = labels
	}
}

// WithLevel allows to set the logging level.
func WithLevel(level flog.Level) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// WithDecorate allows to a function that modifies the log message before it is written.
func WithDecorate(f func(string) string) Option {
	return func(l *Logger) {
		l.decorate = f
	}
}

// WithOnError allows to set a function that is called when an error is logged.
func WithOnError(f func(name string)) Option {
	return func(l *Logger) {
		l.onError = f
	}
}
