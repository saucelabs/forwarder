// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"net/url"
	"strings"
)

// normalizeURLScheme ensures that the URL starts with the scheme.
func normalizeURLScheme(uri string) string {
	u := uri
	scheme := "http"
	if strings.HasPrefix(u, "://") {
		u = uri[3:]
	}

	if strings.Contains(u, "://") {
		return u
	}

	if strings.HasSuffix(u, ":443") {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, u)
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
