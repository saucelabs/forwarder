// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package httplog

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/google/martian/v3/messageview"
	"github.com/saucelabs/forwarder/middleware"
)

type Mode string

const (
	None    Mode = "none"
	URL     Mode = "url"
	Headers Mode = "headers"
	Body    Mode = "body"
	Errors  Mode = "errors"
)

func (m Mode) String() string {
	return string(m)
}

type Logger struct {
	log  func(format string, args ...interface{})
	mode Mode
}

// NewLogger returns a logger that logs HTTP requests and responses.
func NewLogger(logFunc func(format string, args ...interface{}), mode Mode) *Logger {
	return &Logger{
		log:  logFunc,
		mode: mode,
	}
}

func (l *Logger) LogFunc() middleware.Logger {
	switch l.mode {
	case "", None:
		return func(e middleware.LogEntry) {}
	case URL:
		return func(e middleware.LogEntry) {
			var w logWriter
			w.Line(e)
			l.log(w.String())
		}
	case Headers:
		return func(e middleware.LogEntry) {
			var w logWriter
			w.Line(e)
			w.Dump(e)
			l.log(w.String())
		}
	case Body:
		return func(e middleware.LogEntry) {
			w := logWriter{body: true}
			w.Line(e)
			w.Dump(e)
			l.log(w.String())
		}
	case Errors:
		return func(e middleware.LogEntry) {
			if e.Status < http.StatusInternalServerError {
				return
			}

			var w logWriter
			w.Line(e)
			w.Dump(e)
			l.log(w.String())
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

func (w *logWriter) Line(e middleware.LogEntry) {
	fmt.Fprintf(&w.b, "%s %s status=%v duration=%s\n",
		e.Request.Method,
		e.Request.URL.Redacted(),
		e.Status,
		e.Duration,
	)
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
