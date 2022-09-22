package run

import (
	"fmt"
	"net/url"
	"strings"
)

func parseBasicAuth(val string) (*url.Userinfo, error) {
	u, p, ok := strings.Cut(val, ":")
	if !ok {
		return nil, fmt.Errorf("expected user:password")
	}
	return url.UserPassword(u, p), nil
}

func parseProxyURI(val string) (*url.URL, error) {
	return url.ParseRequestURI(val)
}

func parseDNSURI(val string) (*url.URL, error) {
	return url.ParseRequestURI(val)
}
