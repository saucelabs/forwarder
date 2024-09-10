// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type HostPort struct {
	Host string
	Port string
}

func (hp HostPort) Validate() error {
	if hp.Host == "" {
		return errors.New("missing host")
	}
	if hp.Port == "" {
		return errors.New("missing port")
	}

	if !isDomainName(hp.Host) {
		if ip := net.ParseIP(hp.Host); ip == nil {
			return fmt.Errorf("invalid host %q", hp.Host)
		}
	}

	if _, err := strconv.ParseUint(hp.Port, 10, 16); err != nil {
		return fmt.Errorf("invalid port %q", hp.Port)
	}

	return nil
}

type HostPortUser struct {
	HostPort
	*url.Userinfo
}

// ParseHostPortUser parses a user:password@host:port string into HostUser.
func ParseHostPortUser(val string) (*HostPortUser, error) {
	if val == "" || !strings.Contains(val, "@") {
		return nil, errors.New("expected user[:password]@host:port")
	}

	idx := strings.LastIndex(val, "@")

	up := val[:idx]
	hp := val[idx:]

	ui, err := ParseUserinfo(up)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse("http://" + wildcardPortTo0(hp))
	if err != nil {
		return nil, err
	}

	hpi := &HostPortUser{
		HostPort: HostPort{
			Host: u.Hostname(),
			Port: u.Port(),
		},
		Userinfo: ui,
	}
	if err := hpi.Validate(); err != nil {
		return nil, err
	}

	return hpi, nil
}

func (hpu *HostPortUser) Validate() error {
	if hpu.Host == "" {
		return errors.New("missing host")
	}
	if hpu.Port == "" {
		return errors.New("missing port")
	}
	if hpu.Userinfo == nil {
		return errors.New("missing user")
	}
	return validatedUserInfo(hpu.Userinfo)
}

func (hpu *HostPortUser) String() string {
	if hpu == nil {
		return ""
	}

	port := hpu.Port
	if port == "0" {
		port = "*"
	}

	p, ok := hpu.Password()
	if !ok {
		return fmt.Sprintf("%s@%s:%s", hpu.Username(), hpu.Host, port)
	}

	return fmt.Sprintf("%s:%s@%s:%s", hpu.Username(), p, hpu.Host, port)
}

func RedactHostPortUser(hpu *HostPortUser) string {
	if hpu == nil {
		return ""
	}

	port := hpu.Port
	if port == "0" {
		port = "*"
	}

	if _, ok := hpu.Password(); !ok {
		return fmt.Sprintf("%s@%s:%s", hpu.Username(), hpu.Host, port)
	}

	return fmt.Sprintf("%s:xxxxx@%s:%s", hpu.Username(), hpu.Host, port)
}
