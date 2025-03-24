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

	"github.com/saucelabs/forwarder/internal/martian"
	"github.com/saucelabs/forwarder/internal/martian/messageview"
	"github.com/saucelabs/forwarder/middleware"
)

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
	if trace := martian.ContextTraceID(e.Request.Context()); trace != "" {
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
	mv.SkipBody(!w.body || (e.Request.Method == http.MethodConnect && e.Status/100 == 2))

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
