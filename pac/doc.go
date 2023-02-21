// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package pac provides a PAC file parser and evaluator.
// Under the hood uses Goja JavaScript VM to run the PAC script.
// It supports Mozilla FindProxyForURL and the Microsoft IPv6 extension FindProxyForURLEx as well as all the helper functions as described in the PAC specification.
package pac

import (
	"regexp"
	"sort"
)

var jsFunctionRegex = regexp.MustCompile(`function\s+([a-zA-Z0-9_]+)\s*\(`)

// SupportedFunctions returns a list of supported javascript functions from the PAC specification.
func SupportedFunctions() []string {
	var all []string //nolint:prealloc // not worth it
	for _, m := range jsFunctionRegex.FindAllStringSubmatch(asciiPacUtilsScript, -1) {
		all = append(all, m[1])
	}

	// Add built-in functions.
	all = append(all,
		"dnsResolve",
		"myIpAddress",
		// IPv6
		"isResolvableEx",
		"isInNetEx",
		"dnsResolveEx",
		"myIpAddressEx",
		"sortIpAddressList",
		"getClientVersion",
		// Alert
		"alert",
	)

	sort.Strings(all)

	return all
}
