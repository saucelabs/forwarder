// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpx

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestServeUnixSocket(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	socketPath, err := os.CreateTemp(t.TempDir(), "socket")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(socketPath.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := ServeUnixSocket(ctx, h, socketPath.Name()); err != nil {
			t.Errorf("serve on Unix socket: %v", err)
		}
	}()

	c := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath.Name())
			},
		},
	}

	// Wait for the server to start.
	for {
		conn, err := net.Dial("unix", socketPath.Name())
		if err == nil {
			conn.Close()
			break
		}
	}

	resp, err := c.Get("http://unix" + socketPath.Name())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	cancel()

	// Wait for the server to stop.

	for {
		time.Sleep(500 * time.Millisecond)
		if _, err := os.Stat(socketPath.Name()); os.IsNotExist(err) {
			break
		}
	}

	_, err = c.Get("http://unix" + socketPath.Name())
	if err == nil {
		t.Error("get: expected an error")
	}
}
