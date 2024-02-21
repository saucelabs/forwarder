// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httplog

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/saucelabs/forwarder/internal/martian"
	"github.com/saucelabs/forwarder/internal/martian/messageview"
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
	case "", None:
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

type logWriter struct {
	b    bytes.Buffer
	body bool
}

func (w *logWriter) String() string {
	return w.b.String()
}

func (w *logWriter) URLLine(e middleware.LogEntry) {
	w.trace(e)
	fmt.Fprintf(&w.b, "%s %s status=%v duration=%s\n",
		e.Request.Method,
		e.Request.URL.Redacted(),
		e.Status,
		e.Duration,
	)
}

func (w *logWriter) ShortURLLine(e middleware.LogEntry) {
	w.trace(e)

	u := e.Request.URL
	scheme, host, path := u.Scheme, u.Host, u.Path
	if scheme != "" {
		scheme += "://"
	}
	if path != "" && path[0] != '/' {
		path = "/" + path
	}

	fmt.Fprintf(&w.b, "%s %s status=%v duration=%s\n",
		e.Request.Method,
		scheme+host+path,
		e.Status,
		e.Duration,
	)
}

func (w *logWriter) trace(e middleware.LogEntry) {
	if trace := martian.TraceID(e.Request.Context()); trace != "" {
		fmt.Fprintf(&w.b, "[%s] ", trace)
	}
}

func (w *logWriter) Dump(e middleware.LogEntry) {
	if err := w.dump(e); err != nil {
		w.error(err)
	}
	w.sep()
}

func (w *logWriter) dump(e middleware.LogEntry) error {
	mv := messageview.New()
	mv.SkipBody(!w.body)

	// Dump request.
	{
		if err := mv.SnapshotRequest(e.Request); err != nil {
			return err
		}
		r, err := mv.Reader()
		if err != nil {
			return err
		}
		if _, err := io.Copy(&w.b, r); err != nil {
			return err
		}
	}

	// Dump response.
	{
		if e.Response == nil {
			return nil
		}
		if err := mv.SnapshotResponse(e.Response); err != nil {
			return err
		}
		r, err := mv.Reader()
		if err != nil {
			return err
		}
		if _, err := io.Copy(&w.b, r); err != nil {
			return err
		}
	}

	return nil
}

func (w *logWriter) error(err error) {
	fmt.Fprintf(&w.b, "\nlogger error: %s\n", err)
}

func (w *logWriter) sep() {
	fmt.Fprint(&w.b, "\n")
}
