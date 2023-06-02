// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package runctx

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// NotifySignals specifies signals that would cause the context to be canceled.
var NotifySignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT} //nolint:gochecknoglobals // This is only useful in main packages.

// Funcs is a list of functions that can be executed in parallel.
type Funcs []func(ctx context.Context) error

// Run executes all funcs in parallel, and returns the first error.
// Function context is canceled when the process receives a signal from NotifySignals.
func (f Funcs) Run() error {
	ctx := context.Background()
	ctx, unregisterSignals := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		unregisterSignals()
		return nil
	})

	for _, fn := range f {
		fn := fn
		eg.Go(func() error { return fn(ctx) })
	}

	return eg.Wait()
}

func Run(fn func(ctx context.Context) error) error {
	return Funcs{fn}.Run()
}
