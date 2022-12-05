//go:build e2e

package e2e

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func BenchmarkRespNoBody(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, *httpbin+"/status/200", http.NoBody)
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
	req, err := http.NewRequest(http.MethodGet, *httpbin+"/stream-bytes/"+fmt.Sprint(n), http.NoBody)
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
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}
