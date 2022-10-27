package forwarder

import "net/url"

type PACResolver interface {
	FindProxyForURL(url *url.URL, hostname string) (string, error)
}
