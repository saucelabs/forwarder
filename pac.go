// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import "net/url"

type PACResolver interface {
	FindProxyForURL(url *url.URL, hostname string) (string, error)
}

type LoggingPACResolver struct {
	Resolver PACResolver
	Logger   Logger
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
