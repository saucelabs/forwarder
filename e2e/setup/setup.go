// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setup

import (
	"fmt"
	"regexp"

	"github.com/saucelabs/forwarder/utils/compose"
)

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

func (r *Runner) runSetup(s *Setup) error {
	cmd, err := compose.NewCommand(s.Compose)
	if err != nil {
		return err
	}

	if err := cmd.Up(); err != nil {
		return fmt.Errorf("compose up: %w", err)
	}

	cb := makeTestCallback(s.Run, r.Debug)()
	cbErr := cb

	if err := cmd.Down(); err != nil {
		return fmt.Errorf("compose down: %w", err)
	}

	if cbErr != nil {
		return fmt.Errorf("test callback: %w", cbErr)
	}

	if !r.Debug {
		if err := cmd.Close(); err != nil {
			return fmt.Errorf("close command: %w", err)
		}
	}

	return nil
}
