// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type Compose struct {
	Services map[string]*Service `yaml:"services,omitempty"`
	Networks map[string]*Network `yaml:"networks,omitempty"`
	Volumes  map[string]*Volume  `yaml:"volumes,omitempty"`
}

func New() *Compose {
	return &Compose{
		Services: make(map[string]*Service),
		Networks: make(map[string]*Network),
		Volumes:  make(map[string]*Volume),
	}
}

func (c *Compose) AddService(s *Service) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if c.Services[s.Name] != nil {
		return fmt.Errorf("service %s already exists", s.Name)
	}

	c.Services[s.Name] = s

	return nil
}

func (c *Compose) AddNetwork(n *Network) error {
	if err := n.Validate(); err != nil {
		return err
	}
	if c.Networks[n.Name] != nil {
		return fmt.Errorf("network %s already exists", n.Name)
	}

	c.Networks[n.Name] = n

	return nil
}

func (c *Compose) AddVolume(v string) error {
	if v == "" {
		return errors.New("volume is required")
	}

	if c.Volumes[v] != nil {
		return nil
	}

	c.Volumes[v] = &Volume{}

	return nil
}

func (c *Compose) WriteTo(w io.Writer) (int, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}
