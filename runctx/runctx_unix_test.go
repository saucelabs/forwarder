// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build unix

package runctx

import (
	"context"
	"syscall"
	"testing"
)

func TestSignal(t *testing.T) {
	g := NewGroup()
	g.Add(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	g.Add(func(ctx context.Context) error {
		return syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	})

	if err := g.Run(); err != nil {
		t.Fatal(err)
	}
}
