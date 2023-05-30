// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"log"
	"regexp"

	"github.com/saucelabs/forwarder/e2e/run/setups"
)

var setup = flag.String("setup", "", "Only run setups matching this regexp")

func setupRegexp() (*regexp.Regexp, error) {
	if *setup == "" {
		return nil, nil //nolint:nilnil // this is intentional
	}
	return regexp.Compile(*setup)
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}
	r, err := setupRegexp()
	if err != nil {
		log.Fatalf("Invalid setup regexp: %v", err)
	}

	for _, s := range setups.All() {
		if r != nil && !r.MatchString(s.Name) {
			continue
		}
		if err := s.Run(); err != nil {
			log.Fatalf("FAIL: %v", err)
		}
	}
	log.Printf("SUCCESS")
}
