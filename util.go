// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"fmt"
	"net/http"
	"net/url"
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

// NormalizeURI ensures that the url has a scheme.
func NormalizeURI(uriToParse string) (*url.URL, error) {
	// Using ParseRequestURI instead of Parse since our use-case is
	// full URLs only. url.ParseRequestURI expects uriToParse to have a scheme.
	localURL, err := url.ParseRequestURI(normalizeURLScheme(uriToParse))
	if err != nil {
		return nil, err
	}
	if localURL.Scheme == "" {
		localURL.Scheme = "http"
	}
	return localURL, nil
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
