// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setup

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/saucelabs/forwarder/utils/compose"
	"golang.org/x/sync/errgroup"
)

type Runner struct {
	Setups        []Setup
	SetupRegexp   *regexp.Regexp
	Decorate      func(*Setup)
	OnComposeUp   func(*Setup)
	OnComposeDown func(*Setup)
	Debug         bool
	Parallel      int

	td errgroup.Group
	mu sync.Mutex
}

func (r *Runner) Run(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)

	p := r.Parallel
	if r.Debug {
		p = 1
	}

	setups := slices.Clone(r.Setups)

	if p != 1 {
		rand.Shuffle(len(setups), func(i, j int) {
			setups[i], setups[j] = setups[j], setups[i]
		})
	}
	if p > 0 {
		g.SetLimit(p)
	}

	if !CI {
		defer func() {
			if err := r.td.Wait(); err != nil {
				fmt.Fprintf(os.Stderr, "teardown error: %v\n", err)
			}
		}()
	}

	for i := range setups {
		if ctx.Err() != nil {
			break
		}

		s := &setups[i]

		if r.SetupRegexp != nil && !r.SetupRegexp.MatchString(s.Name) {
			continue
		}
		if r.Decorate != nil {
			r.Decorate(s)
		}
		g.Go(func() error {
			if err := r.runSetup(s); err != nil {
				return fmt.Errorf("setup %s: %w", s.Name, err)
			}
			return nil
		})
		if r.Debug {
			break
		}
	}

	return g.Wait()
}

func (r *Runner) runSetup(s *Setup) (runErr error) {
	if s.Compose.Services[TestServiceName] == nil {
		return fmt.Errorf("missing %s service", TestServiceName)
	}

	start := time.Now()

	dir := ""
	if r.Debug {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = wd
	}

	var stdout, stderr bytes.Buffer
	cmd, err := compose.NewCommand(s.Compose, dir, &stdout, &stderr)
	if err != nil {
		return err
	}

	if !r.Debug {
		defer func() {
			r.td.Go(func() error {
				if err := cmd.Down("-v"); err != nil {
					return fmt.Errorf("compose down: %w", err)
				}
				if r.OnComposeDown != nil {
					r.OnComposeDown(s)
				}
				if err := cmd.Close(); err != nil {
					return fmt.Errorf("compose close: %w", err)
				}
				return nil
			})
		}()
	}

	defer func() {
		// Protect against concurrent writes to stdout/stderr.
		r.mu.Lock()
		defer r.mu.Unlock()

		if runErr == nil {
			fmt.Fprintf(os.Stdout, "=== setup %s PASS (%s)\n", s.Name, time.Since(start).Round(time.Millisecond))
			return
		}

		w := os.Stderr

		fmt.Fprintf(w, "=== setup %s FAIL (%s)\n", s.Name, time.Since(start).Round(time.Millisecond))

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

		if err := cmd.Logs(); err != nil {
			fmt.Fprintf(w, "failed to get logs: %v\n", err)
		}

		stdout.WriteTo(w)
		stderr.WriteTo(w)
	}()

	// Bring up all services except the test service.
	args := []string{"-d", "--force-recreate", "--remove-orphans"}

	for name := range s.Compose.Services {
		if name == TestServiceName {
			continue
		}
		args = append(args, name)
	}

	if r.OnComposeUp != nil {
		r.OnComposeUp(s)
	}
	if err := cmd.Up(args...); err != nil {
		return fmt.Errorf("compose up: %w", err)
	}

	// Wait for services to be ready.
	waitTimeout := 10 * time.Second
	if CI {
		waitTimeout = 60 * time.Second
	}
	if err := cmd.Wait(time.Second, waitTimeout); err != nil {
		return fmt.Errorf("wait for services: %w", err)
	}

	// Run the test service.
	return cmd.Up("--force-recreate", "--exit-code-from", TestServiceName, TestServiceName)
}
