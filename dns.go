package forwarder

import (
	"context"
	"fmt"
	"net"
	"net/url"
)

// Sets the default DNS resolver by adding a custom dial function.
// The provided DNS addresses are tried one by one until one succeeds.
// Callers MUST ensure that the provided addresses are IPs, otherwise DNS resolution will end up in a loop.
func setupDNS(dnsURIs []*url.URL, d *net.Dialer, log Logger) {
	if len(dnsURIs) == 0 {
		return
	}

	dial := func(ctx context.Context, network, address string) (net.Conn, error) {
		for _, u := range dnsURIs {
			conn, err := d.DialContext(ctx, u.Scheme, u.Host)
			if err != nil {
				log.Errorf("Failed to dial DNS %s: %v", u, err)
				continue
			}
			return conn, nil
		}
		return nil, fmt.Errorf("failed to dial DNS")
	}

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial:     dial,
	}
}
