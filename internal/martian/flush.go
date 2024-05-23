// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package martian

import (
	"bytes"
	"io"
	"mime"
	"net/http"
)

var (
	sseFlushPattern   = [2]byte{'\n', '\n'}
	chunkFlushPattern = [2]byte{'\r', '\n'}
)

func shouldChunk(res *http.Response) bool {
	if res.ProtoMajor != 1 || res.ProtoMinor != 1 {
		return false
	}
	if res.ContentLength != -1 {
		return false
	}

	// Please read 3.3.2 and 3.3.3 of RFC 7230 for more details https://datatracker.ietf.org/doc/html/rfc7230#section-3.3.2.
	if res.Request.Method == http.MethodHead {
		return false
	}
	// The 204/304 response MUST NOT contain a
	// message-body, and thus is always terminated by the first empty line
	// after the header fields.
	if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusNotModified {
		return false
	}
	if res.StatusCode < 200 {
		return false
	}

	return true
}

func isTextEventStream(res *http.Response) bool {
	// The MIME type is defined in https://www.w3.org/TR/eventsource/#text-event-stream
	resCT := res.Header.Get("Content-Type")
	baseCT, _, _ := mime.ParseMediaType(resCT)
	return baseCT == "text/event-stream"
}

type flusher interface {
	Flush() error
}

// patternFlushWriter is an io.Writer that flushes when a pattern is detected.
type patternFlushWriter struct {
	w       io.Writer
	f       flusher
	pattern [2]byte

	last byte
}

func newPatternFlushWriter(w io.Writer, f flusher, pattern [2]byte) *patternFlushWriter {
	return &patternFlushWriter{
		w:       w,
		f:       f,
		pattern: pattern,
	}
}

func (w *patternFlushWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	if err != nil {
		return
	}

	if (w.last == w.pattern[0] && n > 0 && p[0] == w.pattern[1]) || bytes.LastIndex(p, w.pattern[:]) != -1 {
		err = w.f.Flush()
	}

	if n > 0 {
		w.last = p[n-1]
	} else {
		w.last = 0
	}

	return
}
