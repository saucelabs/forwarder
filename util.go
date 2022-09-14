// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// normalizeURLScheme ensures that the URL starts with the scheme.
func normalizeURLScheme(uri string) string {
	uri = strings.TrimSpace(uri)
	uri = strings.TrimPrefix(uri, "://")
	if strings.Contains(uri, "://") {
		return uri
	}

	scheme := "http"
	if strings.HasSuffix(uri, ":443") {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, uri)
}

// IsLocalHost checks whether the destination host is explicitly local host.
// Note: there can be IPv6 addresses it doesn't catch.
func IsLocalHost(req *http.Request) bool {
	localHostIpv4 := regexp.MustCompile(`127\.0\.0\.\d+`)
	hostName := req.URL.Hostname()

	return hostName == "localhost" ||
		localHostIpv4.MatchString(hostName) ||
		hostName == "0:0:0:0:0:0:0:1" ||
		hostName == "::1"
}
