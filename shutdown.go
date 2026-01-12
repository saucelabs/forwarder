// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type shutdownConfig struct {
	ShutdownTimeout time.Duration
	ShutdownSignals []os.Signal
}

func defaultShutdownConfig() shutdownConfig {
	return shutdownConfig{
		ShutdownTimeout: 30 * time.Second,
		ShutdownSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT},
	}
}

func shutdownContext(cfg shutdownConfig) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancels []func()

	if len(cfg.ShutdownSignals) > 0 {
		var cancel context.CancelFunc
		ctx, cancel = signal.NotifyContext(ctx, cfg.ShutdownSignals...)
		cancels = append(cancels, cancel)
	}
	if cfg.ShutdownTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.ShutdownTimeout)
		cancels = append(cancels, cancel)
	}

	return ctx, func() {
		for _, f := range cancels {
			f()
		}
	}
}
