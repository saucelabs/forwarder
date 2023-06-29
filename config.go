// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	_ "unsafe" // for go:linkname

	"golang.org/x/exp/slices"
)

// ParseUserInfo parses a user:password string into *url.Userinfo.
// Username and password cannot be empty.
func ParseUserInfo(val string) (*url.Userinfo, error) {
	if val == "" {
		return nil, nil //nolint:nilnil // nil is a valid value for Userinfo in URL
	}

	u, p, ok := strings.Cut(val, ":")
	if !ok {
		return nil, fmt.Errorf("expected username:password")
	}

	// URL decode the username and password
	u, err := url.QueryUnescape(u)
	if err != nil {
		return nil, err
	}
	p, err = url.QueryUnescape(p)
	if err != nil {
		return nil, err
	}

	ui := url.UserPassword(u, p)
	if err := validatedUserInfo(ui); err != nil {
		return nil, err
	}

	return ui, nil
}

func validatedUserInfo(ui *url.Userinfo) error {
	if ui == nil {
		return nil
	}
	if ui.Username() == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if p, _ := ui.Password(); p == "" {
		return fmt.Errorf("password cannot be empty")
	}

	return nil
}

func wildcardPortTo0(val string) string {
	s := strings.Split(val, ":")
	if s[len(s)-1] == "*" {
		s[len(s)-1] = "0"
	}
	return strings.Join(s, ":")
}

// ParseHostPortUser parses a user:password@host:port string into HostUser.
// User and password cannot be empty.
func ParseHostPortUser(val string) (*HostPortUser, error) {
	u, err := url.Parse("http://" + wildcardPortTo0(val))
	if err != nil {
		return nil, err
	}

	hpi := &HostPortUser{
		Host:     u.Hostname(),
		Port:     u.Port(),
		Userinfo: u.User,
	}
	if err := hpi.Validate(); err != nil {
		return nil, err
	}

	return hpi, nil
}

func ParseProxyURL(val string) (*url.URL, error) {
	scheme, hostport, ok := strings.Cut(val, "://")
	if !ok {
		scheme = "http"
		hostport = val
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   hostport,
	}

	creds, hostport, ok := strings.Cut(hostport, "@")
	if ok {
		ui, err := ParseUserInfo(creds)
		if err != nil {
			return nil, err
		}
		u.User = ui
		u.Host = hostport
	}

	if err := validateProxyURL(u); err != nil {
		return nil, err
	}

	return u, nil
}

func validateProxyURL(u *url.URL) error {
	if u == nil {
		return nil
	}

	{
		supportedSchemes := []string{
			"http",
			"https",
			"socks5",
		}
		if !slices.Contains[string](supportedSchemes, u.Scheme) {
			return fmt.Errorf("unsupported scheme %q, supported schemes are: %s", u.Scheme, strings.Join(supportedSchemes, ", "))
		}
	}

	{
		if err := validatedUserInfo(u.User); err != nil {
			return err
		}
		host := u.Host
		if u.User != nil {
			host = u.User.String() + "@" + host
		}
		c, err := url.Parse(fmt.Sprintf("%s://%s", u.Scheme, host))
		if err != nil {
			return err
		}
		if u.String() != c.String() {
			return fmt.Errorf("unsupported URL elements, format: [<protocol>://]<host>:<port>")
		}
	}

	{
		h := u.Hostname()

		if !isDomainName(h) {
			ip, err := netip.ParseAddr(h)
			if err != nil {
				return fmt.Errorf("IP: %w", err)
			}
			if !ip.IsValid() {
				return fmt.Errorf("IP: %s", ip)
			}
		}
	}

	{
		if u.Port() == "" {
			return fmt.Errorf("port is required")
		}
		p, err := strconv.ParseUint(u.Port(), 10, 16)
		if err != nil {
			return fmt.Errorf("port: %w", err)
		}
		if p == 0 {
			return fmt.Errorf("port cannot be 0")
		}
	}

	return nil
}

func ParseDNSAddress(val string) (netip.AddrPort, error) {
	var empty netip.AddrPort

	host, port, _ := net.SplitHostPort(val)
	if host == "" {
		host = val
	}

	a, err := netip.ParseAddr(host)
	if err != nil {
		return empty, fmt.Errorf("IP: %w", err)
	}

	var p uint16
	if port == "" {
		p = 53
	} else {
		u, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return empty, fmt.Errorf("port: %w", err)
		}
		p = uint16(u)
	}

	ap := netip.AddrPortFrom(a, p)
	if err := validateDNSAddress(ap); err != nil {
		return empty, err
	}

	return ap, nil
}

func validateDNSAddress(p netip.AddrPort) error {
	if !p.IsValid() {
		return fmt.Errorf("IP: %s", p.Addr())
	}
	if p.Port() == 0 {
		return fmt.Errorf("port cannot be 0")
	}
	return nil
}

//go:linkname isDomainName net.isDomainName
func isDomainName(s string) bool

// OpenFileParser returns a parser that calls os.OpenFile.
// If dirPerm is set it will create the directory if it does not exist.
// For empty path the parser returns nil file and nil error.
func OpenFileParser(flag int, perm, dirPerm os.FileMode) func(val string) (*os.File, error) {
	return func(val string) (*os.File, error) {
		if val == "" {
			return nil, nil
		}

		if dirPerm != 0 {
			dir := filepath.Dir(val)
			if err := os.MkdirAll(dir, dirPerm); err != nil {
				return nil, err
			}
		}
		return os.OpenFile(val, flag, perm)
	}
}
