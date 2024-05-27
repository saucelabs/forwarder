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
	"bufio"
	"bytes"
	"io"
	"mime"
	"net/http"
)

func shouldFlush(res *http.Response) bool {
	if res.Request.Method == http.MethodHead {
		return false
	}
	if res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusNotModified {
		return false
	}

	return isTextEventStream(res) || res.ContentLength == -1
}

func isTextEventStream(res *http.Response) bool {
	// The MIME type is defined in https://www.w3.org/TR/eventsource/#text-event-stream
	resCT := res.Header.Get("Content-Type")
	baseCT, _, _ := mime.ParseMediaType(resCT)
	return baseCT == "text/event-stream"
}

// flushAfterChunkWriter works with net/http/internal.chunkedWriter and forces a flush after each chunk is written.
// There is also net/http/internal.FlushAfterChunkWriter that does the same thing nicer, but it is not available.
type flushAfterChunkWriter struct {
	*bufio.Writer
}

func (w flushAfterChunkWriter) WriteString(s string) (n int, err error) {
	n, err = w.Writer.WriteString(s)
	if s == "\r\n" && err == nil {
		err = w.Flush()
	}
	return
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
