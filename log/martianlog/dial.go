// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package martianlog

import (
	"context"
	"net"
	"time"

	martianlog "github.com/saucelabs/forwarder/internal/martian/log"
)

// LoggingDialContext wraps a dial function adding logging.
// This is a temporary solution until we have context-aware logging everywhere.
// It allows us to log the network and address of the connection being established together with the trace ID.
func LoggingDialContext(dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (conn net.Conn, err error) {
		martianlog.Debug(ctx, "opening connection", "network", network, "address", address)

		start := time.Now()
		conn, err = dial(ctx, network, address)
		if err != nil {
			martianlog.Debug(ctx, "failed to establish connection", "network", network, "address", address, "duration", time.Since(start))
		} else {
			martianlog.Debug(ctx, "connection established", "network", network, "address", address, "duration", time.Since(start))
		}

		return
	}
}
