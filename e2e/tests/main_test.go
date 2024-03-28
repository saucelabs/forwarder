// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()
	waitForProxy()
	os.Exit(m.Run())
}

func waitForProxy() {
	for {
		if err := proxyReady(); err != nil {
			fmt.Fprintf(os.Stdout, "waiting for proxy to be ready: %s", err)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
}

func proxyReady() error {
	resp, err := http.Get(proxyAPI + "/readyz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
