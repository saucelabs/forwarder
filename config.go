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
	"regexp"
	"strconv"
	"strings"
	_ "unsafe" // for go:linkname

	"golang.org/x/exp/slices"
)

// ParseUserinfo parses a user:password string into *url.Userinfo.
func ParseUserinfo(val string) (*url.Userinfo, error) {
	if val == "" {
		return nil, fmt.Errorf("expected username[:password]")
	}

	var ui *url.Userinfo
	u, p, ok := strings.Cut(val, ":")
	if !ok {
		ui = url.User(u)
	} else {
		ui = url.UserPassword(u, p)
	}
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
func ParseHostPortUser(val string) (*HostPortUser, error) {
	if val == "" {
		return nil, fmt.Errorf("expected user[:password]@host:port")
	}
	if strings.Index(val, "@") != strings.LastIndex(val, "@") {
		return nil, fmt.Errorf("only one '@' is allowed")
	}

	up, hp, ok := strings.Cut(val, "@")
	if !ok {
		return nil, fmt.Errorf("expected user[:password]@host:port")
	}

	ui, err := ParseUserinfo(up)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse("http://" + wildcardPortTo0(hp))
	if err != nil {
		return nil, err
	}

	hpi := &HostPortUser{
		Host:     u.Hostname(),
		Port:     u.Port(),
		Userinfo: ui,
	}
	if err := hpi.Validate(); err != nil {
		return nil, err
	}

	return hpi, nil
}

func ParseProxyURL(val string) (*url.URL, error) {
	scheme, hpu, ok := strings.Cut(val, "://")
	if !ok {
		scheme = "http"
		hpu = val
	}

	if strings.Index(hpu, "@") != strings.LastIndex(hpu, "@") {
		return nil, fmt.Errorf("only one '@' is allowed")
	}

	var (
		ui  *url.Userinfo
		err error
	)
	up, hp, ok := strings.Cut(hpu, "@")
	if ok {
		ui, err = ParseUserinfo(up)
		if err != nil {
			return nil, err
		}
	} else {
		hp = hpu
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   hp,
		User:   ui,
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
		if !slices.Contains(supportedSchemes, u.Scheme) {
			return fmt.Errorf("unsupported scheme %q, supported schemes are: %s", u.Scheme, strings.Join(supportedSchemes, ", "))
		}
	}

	{
		if err := validatedUserInfo(u.User); err != nil {
			return err
		}

		c, err := url.Parse(fmt.Sprintf("%s://%s", u.Scheme, u.Host))
		if err != nil {
			return err
		}

		uu := *u
		uu.User = nil
		if uu.String() != c.String() {
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

func ParsePrometheusNamespace(val string) (string, error) {
	if err := validatePrometheusNamespace(val); err != nil {
		return "", err
	}

	return val, nil
}

// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
var promNamespaceRegexp = regexp.MustCompile("^[a-zA-Z_:][a-zA-Z0-9_:]*$")

func validatePrometheusNamespace(val string) error {
	if val == "" {
		return nil
	}

	if !promNamespaceRegexp.MatchString(val) {
		return fmt.Errorf("invalid namespace: %s, it must match %q", val, promNamespaceRegexp.String())
	}

	return nil
}
