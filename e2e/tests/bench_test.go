// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"io"
	"net/http"
	"strconv"
	"testing"
)

func BenchmarkRespNoBody(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, httpbin+"/status/200", http.NoBody)
	if err != nil {
		b.Fatal(err)
	}
	tr := newTransport(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := tr.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

func BenchmarkRespBody1k(b *testing.B) {
	benchmarkStreamDataN(b, 1024)
}

func BenchmarkRespBody100k(b *testing.B) {
	benchmarkStreamDataN(b, 100*1024)
}

func benchmarkStreamDataN(b *testing.B, n int64) {
	b.Helper()
	req, err := http.NewRequest(http.MethodGet, httpbin+"/stream-bytes/"+strconv.FormatInt(n, 10), http.NoBody)
	if err != nil {
		b.Fatal(err)
	}
	tr := newTransport(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := tr.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
