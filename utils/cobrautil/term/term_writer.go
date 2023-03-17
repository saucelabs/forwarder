/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package term

import (
	"io"

	"github.com/mitchellh/go-wordwrap"
)

type wordWrapWriter struct {
	limit  uint
	writer io.Writer
}

// NewWordWrapWriter is a Writer that supports a limit of characters on every line
// and does auto word wrapping that respects that limit.
func NewWordWrapWriter(w io.Writer, limit uint) io.Writer {
	return &wordWrapWriter{
		limit:  limit,
		writer: w,
	}
}

func (w wordWrapWriter) Write(p []byte) (nn int, err error) {
	if w.limit == 0 {
		return w.writer.Write(p)
	}
	original := string(p)
	wrapped := wordwrap.WrapString(original, w.limit)
	return w.writer.Write([]byte(wrapped))
}
