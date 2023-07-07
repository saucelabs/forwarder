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
	e2esetup "github.com/saucelabs/forwarder/e2e/setup"
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
		log.Fatalf("invalid setup regexp: %v", err)
	}

	runner := e2esetup.Runner{
		Setups:      setups.All(),
		SetupRegexp: r,
		Decorate: func(s *e2esetup.Setup) {
			log.Printf("running setup %s", s.Name)

			if *debug {
				for _, srv := range s.Compose.Services {
					srv.Environment["FORWARDER_LOG_LEVEL"] = "debug"
					srv.Environment["FORWARDER_LOG_HTTP"] = "headers"
				}
			}
		},
		Debug: *debug,
	}

	if err := runner.Run(); err != nil {
		log.Fatalf(err.Error())
	}

	log.Printf("SUCCESS")
}
