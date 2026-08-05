// Copyright 2026 Sauce Labs Inc., all rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package martian

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestEscapeInvalidPercent(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		changed bool
	}{
		{
			name:    "no percent",
			in:      "GET /foo/bar HTTP/1.1\r\n",
			want:    "GET /foo/bar HTTP/1.1\r\n",
			changed: false,
		},
		{
			name:    "valid escape untouched",
			in:      "GET /foo%20bar HTTP/1.1\r\n",
			want:    "GET /foo%20bar HTTP/1.1\r\n",
			changed: false,
		},
		{
			name:    "bare percent single",
			in:      "GET /foo/%PUBLIC_URL%/bar HTTP/1.1\r\n",
			want:    "GET /foo/%25PUBLIC_URL%25/bar HTTP/1.1\r\n",
			changed: true,
		},
		{
			name:    "absolute URL from proxy client",
			in:      "GET http://google.com/stock-search/%PUBLIC_URL%/favicon.ico HTTP/1.1\r\n",
			want:    "GET http://google.com/stock-search/%25PUBLIC_URL%25/favicon.ico HTTP/1.1\r\n",
			changed: true,
		},
		{
			name:    "percent at end of URL",
			in:      "GET /foo% HTTP/1.1\r\n",
			want:    "GET /foo%25 HTTP/1.1\r\n",
			changed: true,
		},
		{
			name:    "percent with one hex digit",
			in:      "GET /foo%2 HTTP/1.1\r\n",
			want:    "GET /foo%252 HTTP/1.1\r\n",
			changed: true,
		},
		{
			name:    "mixed valid and invalid",
			in:      "GET /a%20b%PUBLIC%/c%2F HTTP/1.1\r\n",
			want:    "GET /a%20b%25PUBLIC%25/c%2F HTTP/1.1\r\n",
			changed: true,
		},
		{
			name:    "hex case-insensitive is valid",
			in:      "GET /foo%aB HTTP/1.1\r\n",
			want:    "GET /foo%aB HTTP/1.1\r\n",
			changed: false,
		},
		{
			name:    "only method no url",
			in:      "GET\r\n",
			want:    "GET\r\n",
			changed: false,
		},
		{
			name:    "multiple spaces around URL",
			in:      "GET     /foo/%PUBLIC_URL%/bar                  HTTP/1.1\r\n",
			want:    "GET     /foo/%25PUBLIC_URL%25/bar                  HTTP/1.1\r\n",
			changed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, changed := escapeInvalidPercent([]byte(tc.in))
			if changed != tc.changed {
				t.Errorf("changed: got %v, want %v", changed, tc.changed)
			}
			if string(got) != tc.want {
				t.Errorf("out:\n  got  %q\n  want %q", got, tc.want)
			}
		})
	}
}

// TestSanitizeRequestLine_ProducesParseableRequest confirms that a request
// line rejected by http.ReadRequest becomes parseable after sanitization,
// and that the URL round-trips to the origin with %25-escaped percents.
func TestSanitizeRequestLine_ProducesParseableRequest(t *testing.T) {
	const raw = "GET http://google.com/stock-search/%PUBLIC_URL%/favicon.ico HTTP/1.1\r\n" +
		"Host: google.com\r\n\r\n"

	// Sanity: stock http.ReadRequest rejects this.
	if _, err := http.ReadRequest(bufio.NewReader(newFakeConn(raw))); err == nil {
		t.Fatal("expected http.ReadRequest to fail on unsanitized input")
	}

	conn := newFakeConn(raw)
	initial := bufio.NewReader(conn)
	// Mirror the real caller: proxyConn does Peek(1) before sanitize, which
	// forces bufio to fill from the underlying reader.
	if _, err := initial.Peek(1); err != nil {
		t.Fatalf("Peek(1): %v", err)
	}
	br, err := sanitizeRequestLine(initial, conn)
	if err != nil {
		t.Fatalf("sanitizeRequestLine: %v", err)
	}

	req, err := http.ReadRequest(br)
	if err != nil {
		t.Fatalf("http.ReadRequest after sanitize: %v", err)
	}

	want := "/stock-search/%25PUBLIC_URL%25/favicon.ico"
	if got := req.URL.RequestURI(); got != want {
		t.Errorf("req.URL.RequestURI() = %q, want %q", got, want)
	}
	if got := req.Host; got != "google.com" {
		t.Errorf("req.Host = %q, want google.com", got)
	}
}

// TestSanitizeRequestLine_NoOpForCleanRequest confirms the fast path leaves
// the reader untouched when there are no bare `%` chars.
func TestSanitizeRequestLine_NoOpForCleanRequest(t *testing.T) {
	const raw = "GET /foo HTTP/1.1\r\nHost: x\r\n\r\n"

	conn := newFakeConn(raw)
	original := bufio.NewReader(conn)
	if _, err := original.Peek(1); err != nil {
		t.Fatalf("Peek(1): %v", err)
	}
	br, err := sanitizeRequestLine(original, conn)
	if err != nil {
		t.Fatalf("sanitizeRequestLine: %v", err)
	}
	if br != original {
		t.Fatal("expected reader to be returned unchanged on clean input")
	}
}

// fakeConn is a net.Conn backed by a byte string for reads and a no-op for
// writes, sufficient for exercising sanitizeRequestLine.
type fakeConn struct {
	r io.Reader
}

func newFakeConn(s string) *fakeConn {
	return &fakeConn{r: newEOFReader(s)}
}

// newEOFReader returns a reader that yields s then EOF. Using strings.NewReader
// works but we want an io.Reader without importing strings just for this.
func newEOFReader(s string) io.Reader { return &stringReader{s: s} }

type stringReader struct {
	s string
	i int
}

func (r *stringReader) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n := copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

func (c *fakeConn) Read(p []byte) (int, error)         { return c.r.Read(p) }
func (c *fakeConn) Write(p []byte) (int, error)        { return len(p), nil }
func (c *fakeConn) Close() error                       { return nil }
func (c *fakeConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *fakeConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *fakeConn) SetDeadline(t time.Time) error      { return nil }
func (c *fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *fakeConn) SetWriteDeadline(t time.Time) error { return nil }
