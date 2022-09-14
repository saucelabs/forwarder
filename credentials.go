// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"strings"

	"github.com/saucelabs/forwarder/internal/logger"
)

type siteCredentialsMatcher struct {
	// host:port credentials for passing basic authentication to requests
	siteCredentials map[string]string

	// host (wildcard port) credentials for passing basic authentication to requests
	siteCredentialsHost map[string]string

	// port (wildcard host) credentials for passing basic authentication to requests
	siteCredentialsPort map[string]string

	// Global wildcard credentials for passing basic authentication to requests
	siteCredentialsWildcard string
}

// match `hostPort` to one of the configured credentials.
// Priority is exact match, then host, then port, then global wildcard.
func (matcher siteCredentialsMatcher) match(hostport string) string {
	if creds, found := matcher.siteCredentials[hostport]; found {
		logger.Get().Tracelnf("Found an auth for %s", hostport)

		return creds
	}

	// hostPort parameter is expected to contain host:port.
	partsLen := 2

	parts := strings.SplitN(hostport, ":", partsLen)

	if len(parts) != partsLen {
		logger.Get().Warnlnf("Unexpected host:port parameter: %s; will not match host or port wildcards", hostport)

		return ""
	}

	host, port := parts[0], parts[1]

	// Host wildcard - check the port only.
	if creds, found := matcher.siteCredentialsPort[port]; found {
		logger.Get().Tracelnf("Found an auth for host wildcard and port match %s", port)

		return creds
	}

	// Port wildcard - check the host only.
	if creds, found := matcher.siteCredentialsHost[host]; found {
		logger.Get().Tracelnf("Found an auth header for port wildcard and host match %s", host)

		return creds
	}

	// Log whether the global wildcard is set.
	// This is a very esoteric use case. It's only added to support a legacy implementation.
	if matcher.siteCredentialsWildcard != "" {
		logger.Get().Traceln("Found an auth for global wildcard")
	}

	return matcher.siteCredentialsWildcard
}

func (matcher siteCredentialsMatcher) isSet() bool {
	if len(matcher.siteCredentials) > 0 ||
		len(matcher.siteCredentialsPort) > 0 ||
		len(matcher.siteCredentialsHost) > 0 ||
		matcher.siteCredentialsWildcard != "" {
		return true
	}

	return false
}
