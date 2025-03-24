// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httplog

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/saucelabs/forwarder/middleware"
)

type Mode string

const (
	None     Mode = "none"
	ShortURL Mode = "short-url"
	URL      Mode = "url"
	Headers  Mode = "headers"
	Body     Mode = "body"
	Errors   Mode = "errors"
)

func (m Mode) String() string {
	if m == "" {
		return DefaultMode.String()
	}
	return string(m)
}

func SplitNameMode(val string) (name string, mode Mode, err error) {
	n, m, ok := strings.Cut(val, ":")
	if ok {
		name = n
		mode = Mode(m)
	} else {
		name = ""
		mode = Mode(val)
	}

	switch mode {
	case None:
		mode = None
	case ShortURL:
		mode = ShortURL
	case URL:
		mode = URL
	case Headers:
		mode = Headers
	case Body:
		mode = Body
	case Errors:
		mode = Errors
	default:
		return "", "", fmt.Errorf("invalid mode %q", mode)
	}

	return
}

var DefaultMode = Errors

type Logger struct {
	log  func(format string, args ...any)
	mode Mode
}

// NewLogger returns a logger that logs HTTP requests and responses.
func NewLogger(logFunc func(format string, args ...any), mode Mode) *Logger {
	if mode == "" {
		mode = DefaultMode
	}
	return &Logger{
		log:  logFunc,
		mode: mode,
	}
}

func (l *Logger) LogFunc() middleware.Logger {
	switch l.mode {
	case None:
		return func(e middleware.LogEntry) {}
	case ShortURL:
		return func(e middleware.LogEntry) {
			var w logWriter
			w.ShortURLLine(e)
			l.log("%s", w.String())
		}
	case URL:
		return func(e middleware.LogEntry) {
			var w logWriter
			w.URLLine(e)
			l.log("%s", w.String())
		}
	case Headers:
		return func(e middleware.LogEntry) {
			var w logWriter
			w.ShortURLLine(e)
			w.Dump(e)
			l.log("%s", w.String())
		}
	case Body:
		return func(e middleware.LogEntry) {
			w := logWriter{body: true}
			w.ShortURLLine(e)
			w.Dump(e)
			l.log("%s", w.String())
		}
	case Errors:
		return func(e middleware.LogEntry) {
			if e.Status < http.StatusInternalServerError {
				return
			}

			var w logWriter
			w.ShortURLLine(e)
			w.Dump(e)
			l.log("%s", w.String())
		}
	default:
		panic(fmt.Sprintf("unknown log mode %s", l.mode))
	}
}
