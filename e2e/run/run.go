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

var (
	setup = flag.String("setup", "", "Only run setups matching this regexp")
	debug = flag.Bool("debug", false, "Enables debug logs and preserves containers after running, this will run only the first matching setup")
)

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

		if *debug {
			for _, srv := range s.Services {
				srv.Environment["FORWARDER_LOG_LEVEL"] = "debug"
				srv.Environment["FORWARDER_LOG_HTTP"] = "headers"
			}
			s.Debug = true
		}
		if err := s.Run(*debug); err != nil {
			log.Fatalf("FAIL: %v", err)
		}
		if *debug {
			break
		}
	}
	log.Printf("SUCCESS")
}
