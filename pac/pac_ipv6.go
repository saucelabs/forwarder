// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sort"

	"github.com/dop251/goja"
)

// This file implements the Microsoft extensions to PAC for IPv6.
// See https://learn.microsoft.com/en-us/windows/win32/winhttp/ipv6-aware-proxy-helper-api-definitions

// Handler for "isResolvableEx(host)".
// Returns TRUE if the host is resolvable to a IPv4 or IPv6 address; otherwise, FALSE.
// See https://learn.microsoft.com/en-us/windows/win32/winhttp/isresolvableex
func (pr *ProxyResolver) isResolvableEx(call goja.FunctionCall) goja.Value {
	return pr.vm.ToValue(pr.dnsResolveEx(call).String() != "")
}

// Handler for "isInNetEx(host, cidr);".
// Returns TRUE if the host is in the same subnet; otherwise, FALSE.
// See https://learn.microsoft.com/en-us/windows/win32/winhttp/isinnetex
func (pr *ProxyResolver) isInNetEx(call goja.FunctionCall) goja.Value {
	if isNullOrUndefined(call.Argument(0)) {
		return goja.Null()
	}
	host, ok := asString(call.Argument(0))
	if !ok {
		return pr.vm.ToValue(false)
	}

	if isNullOrUndefined(call.Argument(1)) {
		return goja.Null()
	}
	cidr, ok := asString(call.Argument(1))
	if !ok {
		return pr.vm.ToValue(false)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return pr.vm.ToValue(false)
	}
	_, mask, err := net.ParseCIDR(cidr)
	if err != nil {
		return pr.vm.ToValue(false)
	}

	return pr.vm.ToValue(mask.Contains(ip))
}

// Handler for "dnsResolveEx(host)".
// Returns a semicolon delimited string containing IPv6 and IPv4 addresses or an empty string if host is not resolvable.
// See https://learn.microsoft.com/en-us/windows/win32/winhttp/dnsresolveex
func (pr *ProxyResolver) dnsResolveEx(call goja.FunctionCall) goja.Value {
	if isNullOrUndefined(call.Argument(0)) {
		return goja.Null()
	}
	host, ok := asString(call.Argument(0))
	if !ok {
		return pr.vm.ToValue(false)
	}

	lookupIP := pr.config.testingLookupIP
	if lookupIP == nil {
		lookupIP = pr.resolver.LookupIP
	}
	ips, err := lookupIP(context.Background(), "ip", host)
	if err != nil {
		return pr.vm.ToValue("")
	}

	return pr.vm.ToValue(semicolonDelimitedString(ips))
}

// Handler for "myIpAddressEx()".
// Returns a semicolon delimited string containing all IP addresses for localhost (IPv6 and/or IPv4), or an empty string if unable to resolve localhost to an IP address.
// See https://learn.microsoft.com/en-us/windows/win32/winhttp/myipaddressex
func (pr *ProxyResolver) myIPAddressEx(_ goja.FunctionCall) goja.Value {
	var ips []net.IP
	if pr.config.testingMyIPAddressEx != nil {
		ips = pr.config.testingMyIPAddressEx
	} else {
		ips = myIPAddress(true)
	}

	if len(ips) == 0 {
		return pr.vm.ToValue("")
	}

	return pr.vm.ToValue(semicolonDelimitedString(ips))
}

type parsedIP struct {
	net.IP
	orig string
}

func (p parsedIP) String() string {
	return p.orig
}

// Handler for "sortIpAddressList(ipAddressList)".
// Returns a list of sorted semicolon delimited IP addresses or an empty string if unable to sort the IP Address list.
// If the IP Address list contains both IPv4 and IPv6 addresses, the IPv6 addresses are returned first.
// https://learn.microsoft.com/en-us/windows/win32/winhttp/sortipaddresslist
func (pr *ProxyResolver) sortIPAddressList(call goja.FunctionCall) goja.Value {
	if isNullOrUndefined(call.Argument(0)) {
		return goja.Null()
	}
	s, ok := asString(call.Argument(0))
	if !ok {
		return pr.vm.ToValue(false)
	}

	ips, err := asSlice(s, ";", func(v string) (parsedIP, error) {
		ip := net.ParseIP(v)
		if ip == nil {
			return parsedIP{}, fmt.Errorf("invalid IP address")
		}
		return parsedIP{
			IP:   ip,
			orig: v,
		}, nil
	})
	if err != nil {
		return pr.vm.ToValue(false)
	}
	if len(ips) == 0 {
		return pr.vm.ToValue(false)
	}

	sort.Slice(ips, func(i, j int) bool {
		// Compare IPv4 to IPv4 and IPv6 to IPv6
		if (ips[i].To4() != nil) == (ips[j].To4() != nil) {
			return bytes.Compare(ips[i].IP, ips[j].IP) < 0
		}
		// Put IPv6 addresses first.
		return ips[i].To4() == nil
	})
	return pr.vm.ToValue(semicolonDelimitedString(ips))
}

// Handler for "getClientVersion()".
// Returns the appropriate versions number of the WPAD engine, currently 1.0.
// See https://learn.microsoft.com/en-us/windows/win32/winhttp/getclientversion
func (pr *ProxyResolver) getClientVersion(_ goja.FunctionCall) goja.Value {
	return pr.vm.ToValue("1.0")
}
