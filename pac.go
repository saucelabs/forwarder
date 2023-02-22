// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"net/url"

	"github.com/saucelabs/forwarder/log"
)

type PACResolver interface {
	// FindProxyForURL calls FindProxyForURL or FindProxyForURLEx function in the PAC script.
	// The hostname is optional, if empty it will be extracted from URL.
	FindProxyForURL(url *url.URL, hostname string) (string, error)
}

type LoggingPACResolver struct {
	Resolver PACResolver
	Logger   log.Logger
}

func (r *LoggingPACResolver) FindProxyForURL(u *url.URL, hostname string) (string, error) {
	s, err := r.Resolver.FindProxyForURL(u, hostname)
	if err != nil {
		r.Logger.Errorf("FindProxyForURL(%q, %q) failed: %s", u.Redacted(), hostname, err)
	} else {
		r.Logger.Debugf("FindProxyForURL(%q, %q) -> %q", u.Redacted(), hostname, s)
	}
	return s, err
}
