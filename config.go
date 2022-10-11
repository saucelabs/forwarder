package forwarder

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

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

// ParseProxyURI parser a Proxy URI as URL
//
// Requirements:
// - Protocol: http, https, socks5, socks, quic.
// - Hostname min 4 chars.
// - Port in a valid range: 1 - 65535.
// - (Optional) username and password.
func ParseProxyURI(val string) (*url.URL, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}
	if err := validateProxyURI(u); err != nil {
		return nil, err
	}

	return u, nil
}

const minHostLength = 4

func validateProxyURI(u *url.URL) error {
	if u == nil {
		return nil
	}
	if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "socks5" && u.Scheme != "socks" && u.Scheme != "quic" {
		return fmt.Errorf("invalid scheme %q", u.Scheme)
	}
	if len(u.Hostname()) < minHostLength {
		return fmt.Errorf("invalid hostname: %s is too short", u.Hostname())
	}
	if u.Port() == "" {
		return fmt.Errorf("port is required")
	}
	if !isPort(u.Port()) {
		return fmt.Errorf("invalid port: %s", u.Port())
	}
	if err := validatedUserInfo(u.User); err != nil {
		return err
	}

	return nil
}

// ParseDNSURI parses a DNS URI as URL.
// It supports IP only or full URL.
// Hostname is not allowed.
// Examples: `udp://1.1.1.1:53`, `1.1.1.1`.
//
// Requirements:
// - (Optional) protocol: udp, tcp (default udp)
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
		return fmt.Errorf("invalid protocol: %s, supported protocols are udp and tcp", u.Scheme)
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
