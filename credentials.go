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

	"github.com/saucelabs/forwarder/log"
)

type CredentialsMatcher struct {
	hostport map[string]*url.Userinfo
	host     map[string]*url.Userinfo
	port     map[string]*url.Userinfo
	global   *url.Userinfo
	log      log.StructuredLogger
}

func NewCredentialsMatcher(credentials []*HostPortUser, log log.StructuredLogger) (*CredentialsMatcher, error) {
	if len(credentials) == 0 {
		return nil, nil //nolint:nilnil // nil is a valid value
	}

	m := &CredentialsMatcher{
		hostport: make(map[string]*url.Userinfo),
		host:     make(map[string]*url.Userinfo),
		port:     make(map[string]*url.Userinfo),
		log:      log,
	}

	for i, hpu := range credentials {
		withRowInfo := func(err error) error {
			return fmt.Errorf("%w at pos %d", err, i) //nolint:scopelint // false positive
		}

		if err := hpu.Validate(); err != nil {
			return nil, withRowInfo(err)
		}

		switch {
		case hpu.Host == "*" && hpu.Port == "0":
			if m.global != nil {
				return nil, withRowInfo(errors.New("duplicate global input"))
			}
			m.global = hpu.Userinfo
		case hpu.Host == "*":
			if _, ok := m.port[hpu.Port]; ok {
				return nil, withRowInfo(fmt.Errorf("duplicate wildcard host with port %s credentis", hpu.Port))
			}
			m.port[hpu.Port] = hpu.Userinfo
		case hpu.Port == "0":
			if _, ok := m.host[hpu.Host]; ok {
				return nil, withRowInfo(fmt.Errorf("duplicate wildcard port with host %s credentis", hpu.Host))
			}
			m.host[hpu.Host] = hpu.Userinfo
		default:
			hostport := net.JoinHostPort(hpu.Host, hpu.Port)
			if _, ok := m.hostport[hostport]; ok {
				return nil, errors.New("duplicate input")
			}
			m.hostport[hostport] = hpu.Userinfo
		}
	}

	return m, nil
}

// MatchURL adds standard http and https ports if they are missing in URL and calls Match function.
func (m *CredentialsMatcher) MatchURL(u *url.URL) *url.Userinfo {
	if m == nil || u == nil {
		return nil
	}

	const (
		httpPort  = 80
		httpsPort = 443
	)

	hostport := u.Host
	if u.Port() == "" {
		switch u.Scheme {
		case "http":
			hostport = fmt.Sprintf("%s:%d", u.Host, httpPort)
		case "https":
			hostport = fmt.Sprintf("%s:%d", u.Host, httpsPort)
		default:
			m.log.Error("cannot to determine port", "url", u.Redacted())
			return nil
		}
	}

	return m.Match(hostport)
}

// Match `hostport` to one of the configured input.
// Priority is exact Match, then host, then port, then global wildcard.
func (m *CredentialsMatcher) Match(hostport string) *url.Userinfo {
	if m == nil {
		return nil
	}

	if u, ok := m.hostport[hostport]; ok {
		m.log.Debug(hostport)
		return u
	}

	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		m.log.Info("invalid hostport", "hostport", hostport)
		return nil
	}

	// Host wildcard - check the port only.
	if u, ok := m.port[port]; ok {
		m.log.Debug("host=*", "port", port)
		return u
	}

	// Port wildcard - check the host only.
	if u, ok := m.host[host]; ok {
		m.log.Debug("port=*", "host", host)
		return u
	}

	// Log whether the global wildcard is set.
	// This is a very esoteric use case. It's only added to support a legacy implementation.
	if m.global != nil {
		m.log.Debug("global wildcard")
		return m.global
	}

	return nil
}
