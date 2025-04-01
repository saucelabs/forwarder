// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
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
	Logger   log.StructuredLogger
}

func (r *LoggingPACResolver) FindProxyForURL(u *url.URL, hostname string) (string, error) {
	s, err := r.Resolver.FindProxyForURL(u, hostname)
	if err != nil {
		r.Logger.Error("FindProxyForURL failed", "url", u.Redacted(), "hostname", hostname, "error", err)
	} else {
		r.Logger.Debug("FindProxyForURL", "url", u.Redacted(), "hostname", hostname, "result", s)
	}
	return s, err
}
