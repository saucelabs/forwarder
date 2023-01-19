// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package httplog

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/martian/v3/messageview"
	"github.com/saucelabs/forwarder/middleware"
)

type LoggerMode string

const (
	URLLogMode     LoggerMode = "url"
	HeadersLogMode LoggerMode = "headers"
	BodyLogMode    LoggerMode = "body"
	ErrOnlyLogMode LoggerMode = "error"
)

func (m LoggerMode) Validate() error {
	switch m {
	case URLLogMode, HeadersLogMode, BodyLogMode, ErrOnlyLogMode:
		return nil
	}

	return fmt.Errorf("log mode %s not found", m)
}

func ParseMode(val string) (LoggerMode, error) {
	mode := LoggerMode(val)
	switch mode {
	case URLLogMode, HeadersLogMode, BodyLogMode, ErrOnlyLogMode:
		return mode, nil
	}

	return "", fmt.Errorf("log mode %s not found", mode)
}

type Logger struct {
	log  func(format string, args ...interface{})
	mode LoggerMode
}

// NewLogger returns a logger that logs HTTP requests and responses.
func NewLogger(logFunc func(format string, args ...interface{}), mode LoggerMode) *Logger {
	return &Logger{
		log:  logFunc,
		mode: mode,
	}
}

var dash = strings.Repeat("-", 80) //nolint:gomnd,gochecknoglobals // 80 is good

func logURL(w io.Writer, u *url.URL, status int, duration time.Duration) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, dash)
	fmt.Fprintf(w, "Request to %s, status: %v, duration: %s\n", u.Redacted(), status, duration)
	fmt.Fprintln(w, dash)
}

func logEnd(w io.Writer) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, dash)
}

func logError(w io.Writer, err error) {
	fmt.Fprintf(w, "logger error: %v\n", err)
}

func logRequest(w io.Writer, mv *messageview.MessageView, req *http.Request) error {
	if err := mv.SnapshotRequest(req); err != nil {
		return err
	}

	var opts []messageview.Option
	r, err := mv.Reader(opts...)
	if err != nil {
		return err
	}

	io.Copy(w, r) //nolint:errcheck // skip

	return nil
}

func logResponse(w io.Writer, mv *messageview.MessageView, res *http.Response) error {
	if err := mv.SnapshotResponse(res); err != nil {
		return err
	}

	var opts []messageview.Option
	r, err := mv.Reader(opts...)
	if err != nil {
		return err
	}

	io.Copy(w, r) //nolint:errcheck // skip

	return nil
}

func (l *Logger) logURL(e middleware.LogEntry) {
	b := &bytes.Buffer{}
	logURL(b, e.Request.URL, e.Status, e.Duration)
	l.log("%s", b.String())
}

func (l *Logger) logWithHeaders(e middleware.LogEntry) {
	b := &bytes.Buffer{}

	logURL(b, e.Request.URL, e.Status, e.Duration)

	mv := messageview.New()
	mv.SkipBody(true)

	if err := logRequest(b, mv, e.Request); err != nil {
		logError(b, err)
	}

	// It has to be checked whether there is a Response, since current server implementation with HTTP handlers
	// is not capable to pass Response to the entry.
	if e.Response != nil {
		if err := logResponse(b, mv, e.Response); err != nil {
			logError(b, err)
		}
	}

	logEnd(b)
	l.log("%s", b.String())
}

func (l *Logger) logWithBody(e middleware.LogEntry) {
	b := &bytes.Buffer{}

	logURL(b, e.Request.URL, e.Status, e.Duration)

	mv := messageview.New()
	mv.SkipBody(false)

	if err := logRequest(b, mv, e.Request); err != nil {
		logError(b, err)
	}

	// It has to be checked whether there is a Response, since current server implementation with HTTP handlers
	// is not capable to pass Response to the entry.
	if e.Response != nil {
		if err := logResponse(b, mv, e.Response); err != nil {
			logError(b, err)
		}
	}

	logEnd(b)
	l.log("%s", b.String())
}

func (l *Logger) logError(e middleware.LogEntry) {
	if e.Status >= http.StatusInternalServerError {
		l.logWithBody(e)
	}
}

func (l *Logger) LogFunc() middleware.Logger {
	switch l.mode {
	case URLLogMode:
		return l.logURL

	case HeadersLogMode:
		return l.logWithHeaders

	case BodyLogMode:
		return l.logWithBody

	case ErrOnlyLogMode:
		return l.logError
	}

	return nil
}
