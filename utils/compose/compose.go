// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

type Compose struct {
	Path     string              `yaml:"-"`
	Version  string              `yaml:"version"`
	Services map[string]*Service `yaml:"services,omitempty"`
	Networks map[string]*Network `yaml:"networks,omitempty"`
}

func newCompose() *Compose {
	return &Compose{
		Path:     "docker-compose.yaml",
		Version:  "3.8",
		Services: make(map[string]*Service),
		Networks: make(map[string]*Network),
	}
}

func (c *Compose) addService(s *Service) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if c.Services[s.Name] != nil {
		return fmt.Errorf("service %s already exists", s.Name)
	}

	c.Services[s.Name] = s

	return nil
}

func (c *Compose) addNetwork(n *Network) error {
	if err := n.Validate(); err != nil {
		return err
	}
	if c.Networks[n.Name] != nil {
		return fmt.Errorf("network %s already exists", n.Name)
	}

	c.Networks[n.Name] = n

	return nil
}

func (c *Compose) Run(callback func() error, preserve bool) error {
	if err := c.save(c.Path); err != nil {
		return fmt.Errorf("compose save: %w", err)
	}
	if err := c.up(); err != nil {
		return fmt.Errorf("compose up: %w", err)
	}
	if err := callback(); err != nil {
		return err
	}
	if !preserve {
		if err := c.down(); err != nil {
			return fmt.Errorf("compose down: %w", err)
		}
	}

	return nil
}

func (c *Compose) save(path string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return writeFile(path, b)
}

func (c *Compose) up() error {
	return runQuietly(c.dockerCompose("up", "-d", "--wait", "--force-recreate", "--remove-orphans"))
}

func (c *Compose) down() error {
	return runQuietly(c.dockerCompose("down", "-v", "--remove-orphans"))
}

func (c *Compose) dockerCompose(args ...string) *exec.Cmd {
	return exec.Command("docker-compose", append([]string{ //nolint:gosec // G204: Subprocess launched with a potential tainted input or cmd arguments
		"-f", c.Path,
	}, args...)...)
}

func runQuietly(cmd *exec.Cmd) error {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stdout.WriteTo(os.Stdout)
		stderr.WriteTo(os.Stderr)
	}
	return err
}
