// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package e2e

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func BenchmarkRespNoBody(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, testConfig.HTTPBin+"/status/200", http.NoBody)
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
	req, err := http.NewRequest(http.MethodGet, testConfig.HTTPBin+"/stream-bytes/"+fmt.Sprint(n), http.NoBody)
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
