package forwarder

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

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

func (c *ProxyConfig) Clone() *ProxyConfig {
	v := new(ProxyConfig)
	deepCopy(v, c)
	return v
}

func (c *ProxyConfig) Validate() error {
	v := validation.Validator()
	return v.Struct(c)
}

// ParseUserInfo parses a user:password string into *url.Userinfo.
// Username and password cannot be empty.
func ParseUserInfo(val string) (*url.Userinfo, error) {
	if val == "" {
		return nil, nil //nolint:nilnil // nil is a valid value for Userinfo in URL
	}

	u, p, ok := strings.Cut(val, ":")
	if !ok {
		return nil, fmt.Errorf("expected username:password")
	}
	ui := url.UserPassword(u, p)
	if err := validatedUserInfo(ui); err != nil {
		return nil, err
	}

	return ui, nil
}

func validatedUserInfo(ui *url.Userinfo) error {
	if ui == nil {
		return nil
	}
	if ui.Username() == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if p, _ := ui.Password(); p == "" {
		return fmt.Errorf("password cannot be empty")
	}

	return nil
}

// ParseDNSURI parses a DNS URI as URL.
// It supports IP only or full URL.
// Hostname is not allowed.
// Examples: `udp://1.1.1.1:53`, `1.1.1.1`.
//
// Requirements:
// - (Optional) protocol: udp, tcp (default upd)
// - Only IP not a hostname.
// - (Optional) port in a valid range: 1 - 65535 (default 53).
// - No username and password.
// - No path, query, and fragment.
func ParseDNSURI(val string) (*url.URL, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		*u = url.URL{Host: val}
	}
	if u.Scheme == "" {
		u.Scheme = "udp"
	}
	if u.Port() == "" {
		u.Host += ":53"
	}
	if err := validateDNSURI(u); err != nil {
		return nil, err
	}

	return u, nil
}

func validateDNSURI(u *url.URL) error {
	if u.Scheme != "udp" && u.Scheme != "tcp" {
		return fmt.Errorf("invalid protocol: %s, supported protocols are upd and tcp", u.Scheme)
	}
	if net.ParseIP(u.Hostname()) == nil {
		return fmt.Errorf("invalid hostname: %s DNS must be an IP address", u.Hostname())
	}
	if !isPort(u.Port()) {
		return fmt.Errorf("invalid port: %s", u.Port())
	}
	if u.User != nil {
		return fmt.Errorf("username and password are not allowed in DNS URI")
	}
	if u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("path, query, and fragment are not allowed in DNS URI")
	}

	return nil
}

// isPort returns true iff port string is a valid port number.
func isPort(port string) bool {
	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	return p >= 1 && p <= 65535
}
