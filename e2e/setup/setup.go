// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setup

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/saucelabs/forwarder/utils/compose"
)

const TestServiceName = "test"

type Setup struct {
	Name    string
	Compose *compose.Compose
	Run     string
}

type Runner struct {
	Setups      []Setup
	SetupRegexp *regexp.Regexp
	Decorate    func(*Setup)
	Debug       bool
}

func (r *Runner) Run() error {
	for i := range r.Setups {
		s := &r.Setups[i]

		if r.SetupRegexp != nil && !r.SetupRegexp.MatchString(s.Name) {
			continue
		}
		if r.Decorate != nil {
			r.Decorate(s)
		}
		if err := r.runSetup(s); err != nil {
			return err
		}
		if r.Debug {
			break
		}
	}

	return nil
}

func (r *Runner) runSetup(s *Setup) (runErr error) {
	cmd, err := compose.NewCommand(s.Compose)
	if err != nil {
		return err
	}

	defer func() {
		if runErr != nil {
			w := os.Stderr

			fmt.Fprintf(w, "%s\n", cmd.Dir())

			if b, err := os.ReadFile(cmd.File()); err != nil {
				fmt.Fprintf(w, "failed to read compose file: %v\n", err)
			} else {
				fmt.Fprintf(w, "\n%s\n", b)
			}

			fmt.Fprintf(w, "\n")

			if err := cmd.Ps(); err != nil {
				fmt.Fprintf(w, "failed to ps: %v\n", err)
			}

			fmt.Fprintf(w, "\n")

			var args []string
			for name := range s.Compose.Services {
				if name == TestServiceName {
					continue
				}
				args = append(args, name)
			}
			if err := cmd.Logs(args...); err != nil {
				fmt.Fprintf(w, "failed to get logs: %v\n", err)
			}
		}
	}()

	// Bring up all services except the test service.
	{
		args := []string{"-d", "--force-recreate", "--remove-orphans"}

		for name := range s.Compose.Services {
			if name == TestServiceName {
				continue
			}
			args = append(args, name)
		}

		if err := cmd.Up(args...); err != nil {
			return fmt.Errorf("compose up: %w", err)
		}

		waitTimeout := 10 * time.Second
		if _, ok := os.LookupEnv("CI"); ok {
			waitTimeout = 30 * time.Second
		}
		if err := cmd.Wait(time.Second, waitTimeout); err != nil {
			return fmt.Errorf("wait for services: %w", err)
		}
	}

	// Run the test service.
	{
		args := []string{"--force-recreate", "--exit-code-from", TestServiceName, TestServiceName}

		if err := cmd.Up(args...); err != nil {
			return err
		}
	}

	// Clean up.
	if r.Debug {
		w := os.Stderr
		fmt.Fprintf(w, "left running at %s\n", cmd.Dir())
	} else {
		if err := cmd.Down("-v", "--timeout", "1"); err != nil {
			return fmt.Errorf("compose down: %w", err)
		}
		if err := cmd.Close(); err != nil {
			return fmt.Errorf("close command: %w", err)
		}
	}

	return nil
}
