// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func NewRotatableFile(f *os.File) *RotatableFile {
	w := &RotatableFile{
		ch: make(chan os.Signal, 1),
	}
	w.f.Store(f)
	go w.reopenOnSIGHUP()
	return w
}

type RotatableFile struct {
	f    atomic.Pointer[os.File]
	ch   chan os.Signal
	once sync.Once
}

func (w *RotatableFile) file() *os.File {
	return w.f.Load()
}

func (w *RotatableFile) Write(p []byte) (n int, err error) {
	return w.file().Write(p)
}

var closeSignal = syscall.Signal(-1)

func (w *RotatableFile) Close() error {
	w.once.Do(func() {
		w.ch <- closeSignal
	})
	return w.file().Close()
}

func (w *RotatableFile) Reopen() error {
	nf, err := os.OpenFile(w.file().Name(), DefaultFileFlags, DefaultFileMode)
	if err != nil {
		return err
	}
	f := w.f.Swap(nf)

	time.AfterFunc(5*time.Second, func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close old log file: %v\n", err)
		}
	})

	return nil
}

func (w *RotatableFile) reopenOnSIGHUP() {
	signal.Notify(w.ch, syscall.SIGHUP)
	defer signal.Stop(w.ch)

	for s := range w.ch {
		if s == closeSignal {
			return
		}

		if err := w.Reopen(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to rotate log file: %v\n", err)
		}
	}
}
