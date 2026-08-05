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
	"bytes"
	"io"
	"net"
)

// sanitizeRequestLine peeks the first request line from br and, if the URL
// portion contains bare `%` characters that don't form a valid %XX escape
// (e.g. "%PUBLIC_URL%"), rewrites them to `%25` so http.ReadRequest can parse
// the URL. Origin servers decode `%25` back to `%`, so the request semantics
// are preserved end-to-end.
func sanitizeRequestLine(br *bufio.Reader, conn net.Conn) (*bufio.Reader, error) {
	line, ok := peekRequestLine(br)
	if !ok {
		return br, nil
	}
	fixed, changed := escapeInvalidPercent(line)
	if !changed {
		return br, nil
	}

	if _, err := br.Discard(len(line)); err != nil {
		return br, err
	}
	rest, _ := br.Peek(br.Buffered())

	prefix := make([]byte, 0, len(fixed)+len(rest))
	prefix = append(prefix, fixed...)
	prefix = append(prefix, rest...)

	return bufio.NewReader(io.MultiReader(bytes.NewReader(prefix), conn)), nil
}

// peekRequestLine returns the first line (up to and including \n) from br
// without consuming it, using only bytes already buffered. Returns
// (nil, false) if the line does not end within the currently buffered data.
//
// Peeking only what's buffered avoids blocking on a socket read that may
// never return (e.g. an idle keep-alive connection waiting for the next
// request). In the common case, the caller has already triggered a fill via
// bufio.Reader.Peek(1) before this is called, and the whole request line
// arrives in a single Read, so it will be in the buffer.
func peekRequestLine(br *bufio.Reader) ([]byte, bool) {
	buf, _ := br.Peek(br.Buffered())
	if i := bytes.IndexByte(buf, '\n'); i >= 0 {
		return buf[:i+1], true
	}
	return nil, false
}

// escapeInvalidPercent rewrites the URL slice of an HTTP/1.x request line,
// replacing any `%` not followed by two hex digits with `%25`. Returns the
// (possibly reallocated) line and whether any substitution was made.
func escapeInvalidPercent(line []byte) ([]byte, bool) {
	if !hasBarePercent(line) {
		return line, false
	}
	out := make([]byte, 0, len(line)+8)
	for i := range len(line) {
		if line[i] == '%' && !validEscape(line, i) {
			out = append(out, '%', '2', '5')
			continue
		}
		out = append(out, line[i])
	}
	return out, true
}

func hasBarePercent(s []byte) bool {
	for i := range len(s) {
		if s[i] == '%' && !validEscape(s, i) {
			return true
		}
	}
	return false
}

func validEscape(s []byte, i int) bool {
	return i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2])
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}
