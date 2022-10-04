package forwarder

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestResolverLookupHost(t *testing.T) {
	c := &DNSConfig{
		Servers: []*url.URL{{Scheme: "udp", Host: "1.1.1.1:53"}},
		Timeout: 5 * time.Second,
	}
	r, err := NewResolver(c, stdLogger{})
	if err != nil {
		t.Fatal(err)
	}

	addr, err := r.LookupHost(context.Background(), "google.com")
	if err != nil {
		t.Errorf("LookupHost failed: %v", err)
	}
	if len(addr) == 0 {
		t.Errorf("LookupHost returned no address")
	}
}
