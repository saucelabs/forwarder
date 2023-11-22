// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package runctx

import (
	"errors"
	"testing"

	"golang.org/x/net/context"
)

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := NewGroup()
	g.Add(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	g.Add(func(ctx context.Context) error {
		cancel()
		return nil
	})

	if err := g.RunContext(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestError(t *testing.T) {
	testErr := errors.New("test")

	g := NewGroup()
	g.Add(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	g.Add(func(ctx context.Context) error {
		return testErr
	})

	if err := g.Run(); err != testErr { //nolint:errorlint // test it explicitly
		t.Fatal(err)
	}
}
