// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build windows

package runctx

import (
	"context"
	"syscall"
	"testing"
)

func sendCtrlBreak(t *testing.T, pid int) {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		t.Fatalf("LoadDLL: %v\n", e)
	}
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		t.Fatalf("FindProc: %v\n", e)
	}
	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		t.Fatalf("GenerateConsoleCtrlEvent: %v\n", e)
	}
}

func TestSignal(t *testing.T) {
	g := NewGroup()
	g.Add(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	g.Add(func(ctx context.Context) error {
		sendCtrlBreak(t, syscall.Getpid())
		return nil
	})

	if err := g.Run(); err != nil {
		t.Fatal(err)
	}
}
