// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"log"

	"github.com/saucelabs/forwarder/e2e/run/setups"
)

func main() {
	for _, setup := range setups.All() {
		if err := setup.Run(); err != nil {
			log.Fatalf("FAIL: %v", err)
		}
	}
	log.Printf("SUCCESS")
}
