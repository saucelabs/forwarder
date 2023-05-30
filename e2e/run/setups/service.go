// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setups

import (
	"github.com/saucelabs/forwarder/e2e/compose"
)

const (
	ProxyServiceName    = "proxy"
	UpstreamServiceName = "upstream-proxy"
	HttpbinServiceName  = "httpbin"
	ForwarderImage      = "saucelabs/forwarder:${FORWARDER_VERSION}"
)

func ProxyService(opts ...compose.ServiceOpt) compose.Opt {
	defaultOpts := []compose.ServiceOpt{
		WithProtocol("http"),
		WithAPIAddress(":10000"),
		WithPorts("3128:3128", "10000:10000"),
		WithWaitFunc(func(s *compose.Service) error {
			return WaitForServerReady("http://localhost:10000")
		}),
	}
	opts = append(defaultOpts, opts...)

	return func(c *compose.Compose) {
		c.AddService(ProxyServiceName, ForwarderImage, opts...)
	}
}

func UpstreamService(opts ...compose.ServiceOpt) compose.Opt {
	defaultOpts := []compose.ServiceOpt{
		WithProtocol("http"),
		WithAPIAddress(":10000"),
		WithPorts("10020:10000"),
		WithWaitFunc(func(s *compose.Service) error {
			return WaitForServerReady("http://localhost:10020")
		}),
	}
	opts = append(defaultOpts, opts...)

	return func(c *compose.Compose) {
		c.AddService(UpstreamServiceName, ForwarderImage, opts...)
	}
}

func HttpbinService(opts ...compose.ServiceOpt) compose.Opt {
	defaultOpts := []compose.ServiceOpt{
		WithProtocol("http"),
		WithCommand("httpbin"),
		WithAPIAddress(":10000"),
		WithPorts("10010:10000"),
		WithWaitFunc(func(s *compose.Service) error {
			return WaitForServerReady("http://localhost:10010")
		}),
	}
	opts = append(defaultOpts, opts...)

	return func(c *compose.Compose) {
		c.AddService(HttpbinServiceName, ForwarderImage, opts...)
	}
}
