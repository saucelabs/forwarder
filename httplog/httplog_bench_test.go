// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httplog

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/middleware"
)

var (
	message   string
	arguments []any
)

func dummyLogFunc(msg string, args ...any) {
	message = msg
	arguments = args
}

// benchmarkLogger runs the given logger for a fixed log entry.
func benchmarkLogger(b *testing.B, logFunc middleware.Logger) {
	b.Helper()

	reqBody := "sample request body"
	req, err := http.NewRequest(http.MethodGet, "http://example.com", io.NopCloser(strings.NewReader(reqBody)))
	if err != nil {
		b.Fatal(err)
	}
	req.Header.Add("X-Custom-Request-1", "RequestHeaderValue")
	req.Header.Add("X-Custom-Request-2", "RequestHeaderValue")
	req.Header.Add("X-Custom-Request-3", "RequestHeaderValue")
	req.TransferEncoding = []string{"chunked"}
	req.Trailer = http.Header{"X-Trailer-Request": []string{"TrailerValue"}}

	res := &http.Response{
		StatusCode:       http.StatusOK,
		Status:           "200 OK",
		Proto:            "HTTP/1.1",
		Header:           make(http.Header),
		TransferEncoding: []string{"chunked"},
	}
	res.Header.Add("X-Custom-Response", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-1", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-2", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-3", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-4", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-5", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-6", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-7", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-8", "ResponseHeaderValue")
	res.Header.Add("X-Custom-Response-9", "ResponseHeaderValue")
	resBody := "sample response body"
	res.Body = io.NopCloser(strings.NewReader(resBody))
	res.ContentLength = int64(len(resBody))
	res.Trailer = http.Header{"X-Trailer-Response": []string{"TrailerValue"}}

	entry := middleware.LogEntry{
		Request:  req,
		Response: res,
		Status:   res.StatusCode,
		Duration: 100 * time.Millisecond,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logFunc(entry)
	}
}

// BenchmarkNormalShortURL benchmarks a normal (non-structured) logger in ShortURL mode.
func BenchmarkNormalShortURL(b *testing.B) {
	logger := NewLogger(dummyLogFunc, ShortURL)
	benchmarkLogger(b, logger.LogFunc())
}

// BenchmarkNormalHeaders benchmarks a normal (non-structured) logger in Headers mode.
func BenchmarkNormalHeaders(b *testing.B) {
	logger := NewLogger(dummyLogFunc, Headers)
	benchmarkLogger(b, logger.LogFunc())
}

// BenchmarkNormalBody benchmarks a normal (non-structured) logger in Body mode.
func BenchmarkNormalBody(b *testing.B) {
	logger := NewLogger(dummyLogFunc, Body)
	benchmarkLogger(b, logger.LogFunc())
}

// BenchmarkStructuredShortURL benchmarks a structured logger in ShortURL mode.
func BenchmarkStructuredShortURL(b *testing.B) {
	logger := NewStructuredLogger(dummyLogFunc, ShortURL)
	benchmarkLogger(b, logger.LogFunc())
}

// BenchmarkStructuredHeaders benchmarks a structured logger in Headers mode.
func BenchmarkStructuredHeaders(b *testing.B) {
	logger := NewStructuredLogger(dummyLogFunc, Headers)
	benchmarkLogger(b, logger.LogFunc())
}

// BenchmarkStructuredBody benchmarks a structured logger in Body mode.
func BenchmarkStructuredBody(b *testing.B) {
	logger := NewStructuredLogger(dummyLogFunc, Body)
	benchmarkLogger(b, logger.LogFunc())
}
