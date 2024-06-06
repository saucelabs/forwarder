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
	"errors"
	"fmt"
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
		b := make([]byte, 64*1024)
		return &b
	},
}

func bicopy(ctx context.Context, cc ...copier) {
	donec := make(chan struct{}, len(cc))
	for i := range cc {
		go cc[i].copy(ctx, donec)
	}
	for range cc {
		<-donec
	}
}

type copier struct {
	name string
	dst  io.Writer
	src  io.Reader
}

func (c copier) copy(ctx context.Context, donec chan<- struct{}) {
	if n, err := c.optimizedCopy(); n >= 0 {
		if err != nil {
			log.Errorf(ctx, "failed to copy %s tunnel: %v", c.name, err)
		} else {
			log.Debugf(ctx, "%s tunnel finished copying", c.name)
		}

		if err := c.closeWrite(); err != nil {
			log.Errorf(ctx, "failed to close write side of %s tunnel: %v", c.name, err)
		}

		donec <- struct{}{}
		return
	}

	bufp := copyBufPool.Get().(*[]byte) //nolint:forcetypeassert // It's *[]byte.
	buf := *bufp
	defer copyBufPool.Put(bufp)

	if err := copyBuffer(c.dst, c.src, buf); err != nil && !isClosedConnError(err) {
		log.Errorf(ctx, "failed to copy %s tunnel: %v", c.name, err)
	} else {
		log.Debugf(ctx, "%s tunnel finished copying", c.name)
	}

	if err := c.closeWrite(); err != nil {
		log.Errorf(ctx, "failed to close write side of %s tunnel: %v", c.name, err)
	}

	donec <- struct{}{}
	return
}

func (c copier) optimizedCopy() (int64, error) {
	// If the reader has a WriteTo method, use it to do the copy.
	// Avoids an allocation and a copy.
	if wt, ok := c.src.(io.WriterTo); ok {
		return wt.WriteTo(c.dst)
	}
	// Similarly, if the writer has a ReadFrom method, use it to do the copy.
	if rt, ok := c.dst.(io.ReaderFrom); ok {
		return rt.ReadFrom(c.src)
	}

	return -1, nil
}

func (c copier) closeWrite() error {
	if cw, ok := asCloseWriter(c.dst); ok {
		return cw.CloseWrite()
	}
	if pw, ok := c.dst.(*io.PipeWriter); ok {
		return pw.Close()
	}

	return fmt.Errorf("half-close not supported by type %T", c.dst)
}

// errInvalidWrite means that a write returned an impossible count.
var errInvalidWrite = errors.New("invalid write result")

type bufnr struct {
	buf []byte
	nr  int
}

type copyWriter struct {
	dst io.Writer
	wc  <-chan bufnr
	rc  chan<- bufnr
}

func (cw copyWriter) loop(errc chan<- error) {
	var err error

	for v := range cw.wc {
		buf, nr := v.buf, v.nr

		nw, ew := cw.dst.Write(buf[0:nr])
		if nw < 0 || nr < nw {
			nw = 0
			if ew == nil {
				ew = errInvalidWrite
			}
		}
		if ew != nil {
			err = ew
			break
		}
		if nr != nw {
			err = io.ErrShortWrite
			break
		}

		cw.rc <- v
	}

	close(cw.rc)
	errc <- err
}

type copyReader struct {
	src io.Reader
	rc  <-chan bufnr
	wc  chan<- bufnr
}

func (cr copyReader) loop(errc chan<- error) {
	var err error

	for v := range cr.rc {
		buf := v.buf

		nr, er := cr.src.Read(buf)
		if nr > 0 {
			v.nr = nr
			cr.wc <- v
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	close(cr.wc)
	errc <- err
}

func copyBuffer(dst io.Writer, src io.Reader, buf []byte) error {
	rc := make(chan bufnr, 2)
	wc := make(chan bufnr, 2)

	rc <- bufnr{buf: buf[:cap(buf)/2]}
	rc <- bufnr{buf: buf[cap(buf)/2:]}

	cw := copyWriter{dst: dst, wc: wc, rc: rc}
	cr := copyReader{src: src, rc: rc, wc: wc}

	errc := make(chan error, 2)
	go cw.loop(errc)
	go cr.loop(errc)

	var err error
	for i := 0; i < 2; i++ {
		e := <-errc
		if e != nil && err == nil {
			err = e
		}
	}
	return err
}
