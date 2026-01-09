// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package compose

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type IPAMConfig struct {
	Subnet  string `yaml:"subnet"`
	Gateway string `yaml:"gateway"`
}

func (c *IPAMConfig) Validate() error {
	if err := validateIPWithMask(c.Subnet); err != nil {
		return fmt.Errorf("network subnet is invalid: %w", err)
	}
	if err := validateIP(c.Gateway); err != nil {
		return fmt.Errorf("network gateway is invalid %w", err)
	}
	return nil
}

type IPAM struct {
	Config []IPAMConfig `yaml:"config"`
}

func (i *IPAM) Validate() error {
	for j := range i.Config {
		if err := i.Config[j].Validate(); err != nil {
			return fmt.Errorf("network IPAM Config[%d] is invalid: %w", j, err)
		}
	}
	return nil
}

type Network struct {
	Name   string `yaml:"name"`
	Driver string `yaml:"driver"`
	IPAM   IPAM   `yaml:"ipam"`
}

func (n *Network) Validate() error {
	if n.Name == "" {
		return errors.New("network name is required")
	}
	if n.Driver == "" {
		return errors.New("network driver is required")
	}
	return n.IPAM.Validate()
}

func validateIP(input string) error {
	if ip := net.ParseIP(input); ip == nil {
		return fmt.Errorf("could not parse IP address: %s", input)
	}

	return nil
}

func validateMask(input string) error {
	mask, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("could not parse mask: %s", input)
	}

	if mask < 0 || mask > 32 {
		return fmt.Errorf("must be between 0 and 32, mask: %d", mask)
	}

	return nil
}

func validateIPWithMask(input string) error {
	ip, mask, ok := strings.Cut(input, "/")
	if !ok {
		return fmt.Errorf("must be in the format of IP/MASK, input: %s", input)
	}
	if err := validateIP(ip); err != nil {
		return err
	}
	return validateMask(mask)
}
