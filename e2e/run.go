// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/saucelabs/forwarder/e2e/prometheus"
	"github.com/saucelabs/forwarder/e2e/setup"
)

var args = struct {
	setup      *string
	prometheus *bool
	debug      *bool
}{
	setup:      flag.String("setup", "", "Only run setups matching this regexp"),
	prometheus: flag.Bool("prom", false, "Add prometheus to the setup"),
	debug:      flag.Bool("debug", false, "Enables debug logs and preserves containers after running, this will run only the first matching setup"),
}

func setupRegexp() (*regexp.Regexp, error) {
	if *args.setup == "" {
		return nil, nil //nolint:nilnil // this is intentional
	}
	return regexp.Compile(*args.setup)
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}
	r, err := setupRegexp()
	if err != nil {
		fmt.Println("invalid setup regexp:", err)
		os.Exit(1)
	}

	runner := setup.Runner{
		Setups:      AllSetups(),
		SetupRegexp: r,
		Decorate: func(s *setup.Setup) {
			fmt.Println("running setup", s.Name)

			if *args.prometheus {
				p := prometheus.Service()
				s.Compose.Services[p.Name] = p
			}

			if *args.debug {
				for _, srv := range s.Compose.Services {
					if strings.HasPrefix(srv.Image, "saucelabs/forwarder") {
						srv.Environment["FORWARDER_LOG_LEVEL"] = "debug"
						srv.Environment["FORWARDER_LOG_HTTP"] = "headers,api:errors"
					}
				}
			}
		},
		Debug: *args.debug,
	}

	if err := runner.Run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("PASS")
}
