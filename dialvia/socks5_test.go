// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dialvia

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"
	"time"
)

func TestSOCKS5ProxyDialer(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		d := SOCKS5Proxy((&net.Dialer{Timeout: 5 * time.Second}).DialContext, &url.URL{Scheme: "socks5", Host: l.Addr().String()})

		donec := make(chan struct{})
		go func() {
			_, err := d.DialContext(ctx, "tcp", "foobar.com:80")
			if !errors.Is(err, context.Canceled) {
				t.Errorf("got %v, want %v", err, context.Canceled)
			}
			close(donec)
		}()

		cancel()
		select {
		case <-time.After(10 * time.Second):
			t.Fatal("timeout")
		case <-donec:
		}
	})
}
