// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func ServeUnixSocket(ctx context.Context, h http.Handler, socketPath string) error {
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove Unix socket %s: %w", socketPath, err)
	}
	defer os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen Unix socket %s: %w", socketPath, err)
	}
	defer l.Close()

	s := http.Server{
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()
		s.Close()
	}()

	err = s.Serve(l)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}
