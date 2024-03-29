// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

	"github.com/saucelabs/forwarder/command/forwarder"
	"go.uber.org/automaxprocs/maxprocs"
)

func main() {
	if _, err := maxprocs.Set(maxprocs.Logger(nil)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set GOMAXPROCS: %v\n", err)
	}
	if _, ok := os.LookupEnv("GOMEMLIMIT"); !ok {
		os.Setenv("GOMEMLIMIT", "250MiB")
	}

	if err := forwarder.Command().Execute(); err != nil {
		os.WriteFile("/dev/termination-log", []byte(err.Error()), 0o644) //nolint // best effort
		os.Exit(1)
	}
}
