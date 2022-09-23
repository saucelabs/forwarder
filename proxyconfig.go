package forwarder

import (
	"net/url"

	"github.com/saucelabs/forwarder/validation"
)

// ProxyConfig definition.
type ProxyConfig struct {
	// LocalProxyURI is the local proxy URI, ex. http://user:password@127.0.0.1:8080.
	// Requirements:
	// - Known schemes: http, https, socks, socks5, or quic.
	// - Hostname or IP.
	// - Port in a valid range: 1 - 65535.
	// - Username and password are optional.
	LocalProxyURI *url.URL `json:"local_proxy_uri"`

	// UpstreamProxyURI is the upstream proxy URI, ex. http://user:password@127.0.0.1:8080.
	// Only one of `UpstreamProxyURI` or `PACURI` can be set.
	// Requirements:
	// - Known schemes: http, https, socks, socks5, or quic.
	// - Hostname or IP.
	// - Port in a valid range: 1 - 65535.
	// - Username and password are optional.
	UpstreamProxyURI *url.URL `json:"upstream_proxy_uri"`

	// PACURI is the PAC URI, which is used to determine the upstream proxy, ex. http://127.0.0.1:8087/data.pac.
	// Only one of `UpstreamProxyURI` or `PACURI` can be set.
	PACURI *url.URL `json:"pac_uri"`

	// Credentials for proxies specified in PAC content.
	PACProxiesCredentials []string `json:"pac_proxies_credentials"`

	// DNSURIs are DNS URIs, ex. udp://1.1.1.1:53.
	// Requirements:
	// - Known schemes: udp, tcp
	// - IP ONLY.
	// - Port in a valid range: 1 - 65535.
	DNSURIs []*url.URL `json:"dns_uris"`

	// ProxyLocalhost if `true`, requests to `localhost`, `127.0.0.*`, `0:0:0:0:0:0:0:1` will be forwarded to upstream.
	ProxyLocalhost bool `json:"proxy_localhost"`

	// SiteCredentials contains URLs with the credentials, ex.:
	// - https://usr1:pwd1@foo.bar:4443
	// - http://usr2:pwd2@bar.foo:8080
	// - usr3:pwd3@bar.foo:8080
	// Proxy will add basic auth headers for requests to these URLs.
	SiteCredentials []string `json:"site_credentials"`
}

func (c *ProxyConfig) Clone() ProxyConfig {
	v := new(ProxyConfig)
	deepCopy(v, c)
	return v
}

func (c *ProxyConfig) Validate() error {
	v := validation.Validator()
	return v.Struct(c)
}
