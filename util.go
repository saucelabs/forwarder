// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"regexp"
	"strings"
)

func deepCopy(dst, src interface{}) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		panic(err)
	}
	if err := gob.NewDecoder(&buf).Decode(dst); err != nil {
		panic(err)
	}
}

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

var localHostIpv4Regexp = regexp.MustCompile(`127\.0\.0\.\d+`)

// isLocalhost checks whether the destination host is explicitly local host.
// Note: there can be IPv6 addresses it doesn't catch.
func isLocalhost(hostName string) bool {
	return hostName == "localhost" ||
		hostName == "0:0:0:0:0:0:0:1" ||
		hostName == "::1" ||
		localHostIpv4Regexp.MatchString(hostName)
}
