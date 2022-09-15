// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"

	"github.com/saucelabs/forwarder/internal/logger"
)

func userInfoBase64(u *url.Userinfo) string {
	return base64.StdEncoding.EncodeToString([]byte(u.String()))
}

type userInfoMatcher struct {
	// host:port input for passing basic authentication to requests
	hostport map[string]*url.Userinfo
	// host (wildcard port) input for passing basic authentication to requests
	host map[string]*url.Userinfo
	// port (wildcard host) input for passing basic authentication to requests
	port map[string]*url.Userinfo
	// Global wildcard input for passing basic authentication to requests
	global *url.Userinfo
}

var nopUserInfoMatcher = (*userInfoMatcher)(nil)

// newUserInfoMatcher takes a list of "user:pass@host:port" strings and creates a matcher.
// Port '0' means a wildcard port.
// Host '*' means a wildcard host.
// Host and port '*:0' will Match everything.
func newUserInfoMatcher(credentials []string) (*userInfoMatcher, error) {
	m := &userInfoMatcher{
		hostport: make(map[string]*url.Userinfo),
		host:     make(map[string]*url.Userinfo),
		port:     make(map[string]*url.Userinfo),
	}
	ok := false

	for i, s := range credentials {
		withRowInfo := func(err error) error {
			return fmt.Errorf("%w at pos %d", err, i) //nolint:scopelint // false positive
		}

		u, err := url.Parse(normalizeURLScheme(s))
		if err != nil {
			return nil, withRowInfo(fmt.Errorf("invalid URL"))
		}
		if u.User.Username() == "" {
			return nil, withRowInfo(fmt.Errorf("missing username"))
		}
		if p, _ := u.User.Password(); p == "" {
			return nil, withRowInfo(fmt.Errorf("missing password"))
		}

		switch {
		case u.Hostname() == "*" && u.Port() == "0":
			if m.global != nil {
				return nil, withRowInfo(fmt.Errorf("duplicate global input"))
			}
			m.global = u.User
			ok = true
		case u.Hostname() == "*":
			if _, ok := m.port[u.Port()]; ok {
				return nil, withRowInfo(fmt.Errorf("duplicate wildcard host with port %s credentis", u.Port()))
			}
			m.port[u.Port()] = u.User
			ok = true
		case u.Port() == "0":
			if _, ok := m.host[u.Hostname()]; ok {
				return nil, withRowInfo(fmt.Errorf("duplicate wildcard port with host %s credentis", u.Hostname()))
			}
			m.host[u.Hostname()] = u.User
			ok = true
		default:
			if _, ok := m.hostport[u.Host]; ok {
				return nil, fmt.Errorf("duplicate input")
			}
			m.hostport[u.Host] = u.User
			ok = true
		}
	}

	if !ok {
		m = nopUserInfoMatcher
	}

	return m, nil
}

// Match `hostport` to one of the configured input.
// Priority is exact Match, then host, then port, then global wildcard.
func (m *userInfoMatcher) Match(hostport string) *url.Userinfo {
	if m == nopUserInfoMatcher {
		return nil
	}

	if u, ok := m.hostport[hostport]; ok {
		logger.Get().Tracelnf("ok an auth for %s", hostport)
		return u
	}

	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		logger.Get().Infof("invalid hostport %s", hostport)
		return nil
	}

	// Host wildcard - check the port only.
	if u, ok := m.port[port]; ok {
		logger.Get().Tracelnf("ok an auth for host wildcard and port Match %s", port)
		return u
	}

	// Port wildcard - check the host only.
	if u, ok := m.host[host]; ok {
		logger.Get().Tracelnf("ok an auth header for port wildcard and host Match %s", host)
		return u
	}

	// Log whether the global wildcard is set.
	// This is a very esoteric use case. It's only added to support a legacy implementation.
	if m.global != nil {
		logger.Get().Traceln("ok an auth for global wildcard")
		return m.global
	}

	return nil
}
