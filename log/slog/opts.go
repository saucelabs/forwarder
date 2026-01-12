// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package slog

// WithOnError allows to set a function that is called when an error is logged.
func WithOnError(f func(name string)) Option {
	return func(l *Logger) {
		l.onError = f
	}
}

// WithAttributes allows to set custom attributes on the logger creation.
func WithAttributes(args ...any) Option {
	return func(l *Logger) {
		l.log = l.log.With(args...)
	}
}
