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
	"regexp"
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

type HostPortPair struct {
	Src, Dst HostPort
}

func (p HostPortPair) String() string {
	return fmt.Sprintf("%s:%s:%s:%s", p.Src.Host, p.Src.Port, p.Dst.Host, p.Dst.Port)
}

func (p HostPortPair) Validate() error {
	if err := p.Src.Validate(); err != nil {
		return fmt.Errorf("src: %w", err)
	}
	if err := p.Dst.Validate(); err != nil {
		return fmt.Errorf("dst: %w", err)
	}

	return nil
}

// ParseHostPortPair parses HOST1:PORT1:HOST2:PORT2 string into HostPortPair.
// HOST1:PORT1 is the source, HOST2:PORT2 is the destination.
func ParseHostPortPair(val string) (HostPortPair, error) {
	const (
		dns = `[.\w\-]+`
		ip4 = `[.0-9]+`
		ip6 = `\[?[:0-9a-fA-F]+\]?`
	)
	hostPortPairRe := regexp.MustCompile(`^(` + dns + `|` + ip4 + `|` + ip6 + `):(\d+):(` + dns + `|` + ip4 + `|` + ip6 + `):(\d+)$`)

	m := hostPortPairRe.FindStringSubmatch(val)
	if m == nil {
		return HostPortPair{}, errors.New("expected src_host:src_port:dst_host:dst_port")
	}

	r := strings.NewReplacer("[", "", "]", "")

	hpp := HostPortPair{
		Src: HostPort{Host: r.Replace(m[1]), Port: m[2]},
		Dst: HostPort{Host: r.Replace(m[3]), Port: m[4]},
	}

	return hpp, hpp.Validate()
}
