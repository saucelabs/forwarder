// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"fmt"
)

type Service struct {
	Name        string            `yaml:"-"`
	Image       string            `yaml:"image,omitempty"`
	Command     string            `yaml:"command,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`

	WaitFunc func() error `yaml:"-"`
}

func (s *Service) Validate() error {
	if s == nil {
		return fmt.Errorf("service is nil")
	}
	if s.Image == "" {
		return fmt.Errorf("service image is empty")
	}
	if s.Name == "" {
		return fmt.Errorf("service name is empty")
	}

	return nil
}

func (s *Service) Wait() error {
	if s.WaitFunc == nil {
		return nil
	}

	if err := s.WaitFunc(); err != nil {
		return fmt.Errorf("service %s: %w", s.Name, err)
	}

	return nil
}
