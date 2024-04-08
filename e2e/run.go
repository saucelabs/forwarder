// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"

	"github.com/saucelabs/forwarder/e2e/forwarder"
	"github.com/saucelabs/forwarder/e2e/setup"
	"github.com/saucelabs/forwarder/utils/compose"
	"golang.org/x/exp/maps"
)

var args = struct {
	setup *string
	run   *string

	debug    *bool
	parallel *int
}{
	setup:    flag.String("setup", "", "Only run setups matching this regexp"),
	run:      flag.String("run", "", "Only run tests matching this regexp"),
	debug:    flag.Bool("debug", false, "Enables debug logs and preserves containers after running, this will run only the first matching setup"),
	parallel: flag.Int("parallel", 1, "How many setups to run in parallel"),
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

	// Protect setups using custom networks from running in parallel.
	var networkMu sync.Mutex

	runner := setup.Runner{
		Setups:      AllSetups(),
		SetupRegexp: r,
		Decorate: func(s *setup.Setup) {
			if *args.debug {
				for _, srv := range s.Compose.Services {
					if strings.HasPrefix(srv.Image, "saucelabs/forwarder") {
						srv.Environment["FORWARDER_LOG_LEVEL"] = "debug"
						srv.Environment["FORWARDER_LOG_HTTP"] = "headers,api:errors"
					}
					if srv.Name == forwarder.ProxyServiceName {
						srv.Ports = append(srv.Ports,
							"3128:3128",
							"10000:10000",
						)
					}
				}
			}

			t := testService(s)
			s.Compose.Services[t.Name] = t
		},
		OnComposeUp: func(s *setup.Setup) {
			if len(s.Compose.Networks) > 0 {
				networkMu.Lock()
			}
		},
		OnComposeDown: func(s *setup.Setup) {
			if len(s.Compose.Networks) > 0 {
				networkMu.Unlock()
			}
		},
		Debug:    *args.debug,
		Parallel: *args.parallel,
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c

		fmt.Println("Stopping...")
		cancel()

		for range c {
			fmt.Println("Waiting for running tests to finish...")
		}
	}()

	if err := runner.Run(ctx); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("PASS")
}

func testService(s *setup.Setup) *compose.Service {
	run := *args.run
	if run == "" {
		run = s.Run
	}

	var cmd bytes.Buffer
	if *args.debug {
		cmd.WriteString("-test.v ")
	}
	fmt.Fprintf(&cmd, "-test.run %q -test.shuffle on -services %s",
		run, strings.Join(maps.Keys(s.Compose.Services), ","))

	c := &compose.Service{
		Name:    setup.TestServiceName,
		Image:   "forwarder-e2e",
		Command: cmd.String(),
	}

	p, ok := s.Compose.Services[forwarder.ProxyServiceName]
	if !ok {
		panic("proxy service not found")
	}
	c.Environment = maps.Clone(p.Environment)

	if h, ok := s.Compose.Services[forwarder.HttpbinServiceName]; ok {
		c.Environment["HTTPBIN_PROTOCOL"] = h.Environment["FORWARDER_PROTOCOL"]
	}

	if len(s.Compose.Networks) > 0 {
		c.Network = map[string]compose.ServiceNetwork{}
		for name := range s.Compose.Networks {
			c.Network[name] = compose.ServiceNetwork{}
		}
	}

	return c
}
