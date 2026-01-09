// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
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
	"fmt"
	"io"
	"sync"
	"time"

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

var bicopyGracefulTimeout = 1 * time.Minute

func bicopy(ctx context.Context, cc ...copier) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	donec := make(chan struct{}, len(cc))
	for i := range cc {
		go cc[i].copy(ctx, donec)
	}

	for i := range cc {
		<-donec
		if i == 0 {
			// Forcibly close all tunnels 1 minute after the first tunnel finished.
			go gracefulCloseAfter(ctx, bicopyGracefulTimeout, cc...)
		}
	}
}

func gracefulCloseAfter(ctx context.Context, d time.Duration, cc ...copier) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(d):
		log.Info(ctx, "forcibly closing tunnel after graceful period", "period", d)
	}
	for i := range cc {
		cc[i].close(ctx)
	}
}

type copier struct {
	name string
	dst  io.Writer
	src  io.Reader
}

func (c copier) copy(ctx context.Context, donec chan<- struct{}) {
	bufp := copyBufPool.Get().(*[]byte) //nolint:forcetypeassert // It's *[]byte.
	buf := *bufp
	defer copyBufPool.Put(bufp)

	if _, err := io.CopyBuffer(c.dst, c.src, buf); err != nil && !isClosedConnError(err) {
		log.Error(ctx, "failed to copy tunnel", "name", c.name, "error", err)
	}
	c.closeWriter(ctx)

	log.Debug(ctx, "tunnel finished copying", "name", c.name)
	donec <- struct{}{}
}

func (c copier) closeWriter(ctx context.Context) {
	var closeErr error
	if cw, ok := asCloseWriter(c.dst); ok {
		closeErr = cw.CloseWrite()
	} else if pw, ok := c.dst.(*io.PipeWriter); ok {
		closeErr = pw.Close()
	} else {
		log.Error(ctx, "cannot close write side of tunnel", "name", c.name, "type", fmt.Sprintf("%T", c.dst))
	}
	if closeErr != nil {
		log.Info(ctx, "failed to close write side of tunnel", "name", c.name, "error", closeErr)
	}
}

func (c copier) close(ctx context.Context) {
	cc, ok := asCloser(c.dst)
	if !ok {
		log.Error(ctx, "cannot close tunnel", "name", c.name, "type", fmt.Sprintf("%T", c.dst))
		return
	}
	if err := cc.Close(); err != nil && !isClosedConnError(err) {
		log.Info(ctx, "failed to close tunnel", "name", c.name, "error", err)
	}
}
