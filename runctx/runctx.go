// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package runctx

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// DefaultNotifySignals specifies signals that would cause the context to be canceled.
var DefaultNotifySignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

// Group is a collection of functions that would be run concurrently.
// The context passed to each function is canceled when any of the signals in NotifySignals is received.
type Group struct {
	NotifySignals []os.Signal
	funcs         []func(ctx context.Context) error
}

func NewGroup(fn ...func(ctx context.Context) error) *Group {
	return &Group{
		funcs: fn,
	}
}

func (g *Group) Add(fn func(ctx context.Context) error) {
	g.funcs = append(g.funcs, fn)
}

// Run runs all the functions concurrently.
// See RunContext for more details.
func (g *Group) Run() error {
	return g.RunContext(context.Background())
}

// RunContext runs all the functions concurrently.
// It returns the first error returned by any of the functions.
// If the context is canceled, it returns nil.
func (g *Group) RunContext(ctx context.Context) error {
	sigs := g.NotifySignals
	if len(sigs) == 0 {
		sigs = DefaultNotifySignals
	}
	ctx, unregisterSignals := signal.NotifyContext(ctx, sigs...)

	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		unregisterSignals()
		return nil
	})

	for _, fn := range g.funcs {
		fn := fn
		eg.Go(func() error { return fn(ctx) })
	}

	err := eg.Wait()
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}
