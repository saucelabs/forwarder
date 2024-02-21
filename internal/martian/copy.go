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
	"context"
	"io"
	"sync"

	"github.com/saucelabs/forwarder/internal/martian/log"
)

func drainBuffer(w io.Writer, r *bufio.Reader) error {
	if n := r.Buffered(); n > 0 {
		rbuf, err := r.Peek(n)
		if err != nil {
			return err
		}
		w.Write(rbuf)
	}
	return nil
}

var copyBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 32*1024)
		return &b
	},
}

func copySync(ctx context.Context, name string, w io.Writer, r io.Reader, donec chan<- bool) {
	bufp := copyBufPool.Get().(*[]byte) //nolint:forcetypeassert // It's *[]byte.
	buf := *bufp
	defer copyBufPool.Put(bufp)

	if _, err := io.CopyBuffer(w, r, buf); err != nil && !isClosedConnError(err) {
		log.Errorf(ctx, "failed to copy %s tunnel: %v", name, err)
	}
	if cw, ok := asCloseWriter(w); ok {
		cw.CloseWrite()
	} else if pw, ok := w.(*io.PipeWriter); ok {
		pw.Close()
	} else {
		log.Errorf(ctx, "cannot close write side of %s tunnel (%T)", name, w)
	}

	log.Debugf(ctx, "%s tunnel finished copying", name)
	donec <- true
}
