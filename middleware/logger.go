// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package middleware

import (
	"net/http"
	"time"

	"github.com/saucelabs/forwarder/internal/martian"
	"github.com/saucelabs/forwarder/utils/httpx"
)

type LogEntry struct {
	Request  *http.Request
	Response *http.Response
	Status   int
	Duration time.Duration
}

func makeLogEntry(req *http.Request, res *http.Response, d time.Duration) LogEntry {
	le := LogEntry{
		Request:  req,
		Response: res,
		Duration: d,
	}
	if res != nil {
		le.Status = res.StatusCode
	}
	return le
}

type Logger func(e LogEntry)

func (l Logger) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		d, ok := w.(delegator)
		if !ok {
			panic("logger middleware requires delegator")
		}
		h.ServeHTTP(w, r)
		l(LogEntry{
			Request:  r,
			Status:   d.Status(),
			Duration: time.Since(start),
		})
	})
}

func (l Logger) WrapRoundTripper(rt http.RoundTripper) http.RoundTripper {
	return httpx.RoundTripperFunc(func(req *http.Request) (res *http.Response, err error) {
		start := time.Now()
		res, err = rt.RoundTrip(req)
		l(makeLogEntry(req, res, time.Since(start)))
		return
	})
}

const startTimeKey = "start-time"

func (l Logger) ModifyRequest(req *http.Request) error {
	ctx := martian.NewContext(req)
	ctx.Set(startTimeKey, time.Now())
	return nil
}

func (l Logger) ModifyResponse(res *http.Response) error {
	ctx := martian.NewContext(res.Request)

	var d time.Duration
	if s, ok := ctx.Get(startTimeKey); ok {
		if ss, ok := s.(time.Time); ok {
			d = time.Since(ss)
		}
	}
	l(makeLogEntry(res.Request, res, d))

	return nil
}
