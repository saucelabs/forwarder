// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"

	"github.com/dop251/goja"
	"golang.org/x/exp/utf8string"
)

type ProxyResolverConfig struct {
	Script    string
	AlertSink io.Writer

	testingLookupIP      func(ctx context.Context, network, host string) ([]net.IP, error)
	testingMyIPAddress   []net.IP
	testingMyIPAddressEx []net.IP
}

func (c *ProxyResolverConfig) Validate() error {
	if c.Script == "" {
		return fmt.Errorf("PAC script is empty")
	}
	return nil
}

// ProxyResolver is a PAC resolver.
// It can be used to resolve a proxy for a given URL.
// It supports both FindProxyForURL and FindProxyForURLEx functions.
// It is not safe for concurrent use.
type ProxyResolver struct {
	config   ProxyResolverConfig
	vm       *goja.Runtime
	fn       goja.Callable
	resolver *net.Resolver
}

// Option allows to set additional options before evaluating the PAC script.
type Option func(vm *goja.Runtime)

func NewProxyResolver(cfg *ProxyResolverConfig, r *net.Resolver, opts ...Option) (*ProxyResolver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if r == nil {
		r = net.DefaultResolver
	}
	pr := &ProxyResolver{
		config:   *cfg,
		vm:       goja.New(),
		resolver: r,
	}

	// Set helper functions.
	if err := pr.registerFunctions(); err != nil {
		return nil, err
	}
	if _, err := pr.vm.RunString(asciiPacUtilsScript); err != nil {
		panic(err)
	}

	// Set additional options before evaluating the PAC script.
	for _, opt := range opts {
		(opt)(pr.vm)
	}

	// Evaluate the PAC script.
	if _, err := pr.vm.RunString(pr.config.Script); err != nil {
		return nil, fmt.Errorf("PAC script: %w", err)
	}

	// Find the FindProxyForURL function.
	fnx, fn := pr.entryPoint()
	if fnx == nil && fn == nil {
		return nil, fmt.Errorf("PAC script: missing required function FindProxyForURL or FindProxyForURLEx")
	}
	if fnx != nil && fn != nil {
		return nil, fmt.Errorf("PAC script: ambiguous entry point, both FindProxyForURL and FindProxyForURLEx are defined")
	}
	if fnx != nil {
		pr.fn = fnx
	} else {
		pr.fn = fn
	}

	return pr, nil
}

func (pr *ProxyResolver) registerFunctions() error {
	helperFn := []struct {
		name string
		fn   func(call goja.FunctionCall) goja.Value
	}{
		// IPv4
		{"dnsResolve", pr.dnsResolve},
		{"myIpAddress", pr.myIPAddress},
		// IPv6
		{"isResolvableEx", pr.isResolvableEx},
		{"isInNetEx", pr.isInNetEx},
		{"dnsResolveEx", pr.dnsResolveEx},
		{"myIpAddressEx", pr.myIPAddressEx},
		{"sortIpAddressList", pr.sortIPAddressList},
		{"getClientVersion", pr.getClientVersion},
		// Alert
		{"alert", pr.alert},
	}
	for _, v := range helperFn {
		if err := pr.vm.Set(v.name, v.fn); err != nil {
			return fmt.Errorf("failed to set helper function %s: %w", v.name, err)
		}
	}

	return nil
}

func (pr *ProxyResolver) alert(call goja.FunctionCall) goja.Value {
	if pr.config.AlertSink != nil {
		fmt.Fprintln(pr.config.AlertSink, "alert:", call.Argument(0).String())
	}
	return goja.Undefined()
}

func (pr *ProxyResolver) entryPoint() (fnx, fn goja.Callable) {
	fnx, _ = goja.AssertFunction(pr.vm.Get("FindProxyForURLEx"))
	fn, _ = goja.AssertFunction(pr.vm.Get("FindProxyForURL"))
	return
}

// FindProxyForURL calls FindProxyForURL or FindProxyForURLEx function in the PAC script.
// The hostname is optional, if empty it will be extracted from URL.
func (pr *ProxyResolver) FindProxyForURL(u *url.URL, hostname string) (string, error) {
	if hostname == "" {
		hostname = u.Hostname()
	}

	v, err := pr.fn(goja.Undefined(), pr.vm.ToValue(u.String()), pr.vm.ToValue(hostname))
	if err != nil {
		return "", fmt.Errorf("PAC script: %w", err)
	}

	s, ok := asString(v)
	if !ok {
		return "", fmt.Errorf("PAC script: unexpected return type %s", v.ExportType())
	}
	if !utf8string.NewString(s).IsASCII() {
		return "", fmt.Errorf("PAC script: non-ASCII characters in the return value %q", s)
	}

	return s, nil
}
