package forwarder

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"
)

type DNSConfig struct {
	// Servers is a list of DNS servers to use, ex. udp://1.1.1.1:53.
	// Requirements:
	// - Known schemes: udp, tcp
	// - IP ONLY.
	// - Port in a valid range: 1 - 65535.
	Servers []*url.URL `json:"servers"`
	// Timeout is the timeout for DNS queries.
	Timeout time.Duration `json:"timeout"`
}

func DefaultDNSConfig() *DNSConfig {
	return &DNSConfig{
		Timeout: 5 * time.Second,
	}
}

func (c *DNSConfig) Validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("no DNS server configured")
	}
	for i, u := range c.Servers {
		if err := validateDNSURI(u); err != nil {
			return fmt.Errorf("servers[%d]: %w", i, err)
		}
	}
	return nil
}

type resolver struct {
	resolver net.Resolver
	dialer   net.Dialer
	servers  []*url.URL
	log      Logger
}

func NewResolver(cfg *DNSConfig, log Logger) (*net.Resolver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log.Infof("Using DNS servers: %s", cfg.Servers)

	r := new(resolver)
	*r = resolver{
		resolver: net.Resolver{
			PreferGo: true,
			Dial:     r.dialDNS,
		},
		dialer: net.Dialer{
			Timeout:  cfg.Timeout,
			Resolver: nopResolver(),
		},
		servers: cfg.Servers,
		log:     log,
	}

	return &r.resolver, nil
}

func nopResolver() *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, fmt.Errorf("no DNS resolver configured")
		},
	}
}

func (r *resolver) dialDNS(ctx context.Context, network, address string) (net.Conn, error) {
	for _, u := range r.servers {
		r.log.Debugf("Dial DNS server %s instead of %s://%s", u.Redacted(), network, address)
		conn, err := r.dialer.DialContext(ctx, u.Scheme, u.Host)
		if err != nil {
			r.log.Errorf("Failed to dial DNS server %s: %v", u, err)
			continue
		}
		return conn, nil
	}

	return nil, fmt.Errorf("failed to dial DNS server")
}
