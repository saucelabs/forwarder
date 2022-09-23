package run

import (
	"net/url"
)

func parseProxyURI(val string) (*url.URL, error) {
	return url.ParseRequestURI(val)
}
