// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"time"

	"github.com/google/martian/v3"
)

type LogEntry struct {
	Request  *http.Request
	Response *http.Response
	Status   int
	Written  int64
	Duration time.Duration
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
			Written:  d.Written(),
			Duration: time.Since(start),
		})
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

	l(LogEntry{
		Request:  res.Request,
		Response: res,
		Status:   res.StatusCode,
		Written:  0, // There seem not to be an easy way of counting it.
		Duration: d,
	})

	return nil
}
