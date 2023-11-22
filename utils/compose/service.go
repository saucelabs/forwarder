// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"fmt"
	"time"
)

type ServiceNetwork struct {
	IPv4 string `yaml:"ipv4_address,omitempty"`
}

type Service struct {
	Name        string                    `yaml:"-"`
	Image       string                    `yaml:"image,omitempty"`
	Command     string                    `yaml:"command,omitempty"`
	Environment map[string]string         `yaml:"environment,omitempty"`
	Ports       []string                  `yaml:"ports,omitempty"`
	Volumes     []string                  `yaml:"volumes,omitempty"`
	HealthCheck *HealthCheck              `yaml:"healthcheck,omitempty"`
	Network     map[string]ServiceNetwork `yaml:"networks,omitempty"`
	Privileged  bool                      `yaml:"privileged,omitempty"`
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

type HealthCheck struct {
	Test []string `yaml:"test,omitempty"`
	// Interval between two health checks, the default is 30 seconds.
	Interval time.Duration `yaml:"interval,omitempty"`
	// The health check command runs the timeout period.
	// If this time is exceeded, the health check is regarded as a failure.
	Timeout time.Duration `yaml:"timeout,omitempty"`
	// When the specified number of consecutive failures, the container status is treated as unhealthy, the default is 3 times.
	Retries uint `yaml:"retries,omitempty"`
	// The number of seconds to start the health check after the container starts, the default is 0 seconds.
	StartPeriod time.Duration `yaml:"start_period,omitempty"`
}
