// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

	"github.com/saucelabs/forwarder/e2e"
	"github.com/saucelabs/forwarder/e2e/testrunner"
)

func main() {
	tests, err := e2e.AllTests()
	if err != nil {
		fmt.Println(err) //nolint:forbidigo // I allow it.
		os.Exit(1)
	}
	for _, t := range tests {
		if err = t.Save("test-data"); err != nil {
			fmt.Println(err) //nolint:forbidigo // I allow it.
			os.Exit(1)
		}
	}

	if err = testrunner.NewRunner(testrunner.RunnerConfig{
		Root:             "test-data",
		ConcurrencyLimit: 3,
	}).Run(); err != nil {
		fmt.Println(err) //nolint:forbidigo // I allow it.
		os.Exit(1)
	}
}
