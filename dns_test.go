// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/log/stdlog"
)

func TestResolverLookupHost(t *testing.T) {
	if _, ok := os.LookupEnv("CI"); ok {
		t.Skip("skipping test in CI environment")
	}

	c := &DNSConfig{
		Servers: []*url.URL{{Scheme: "udp", Host: "1.1.1.1:53"}},
		Timeout: 5 * time.Second,
	}
	r, err := NewResolver(c, stdlog.Default())
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
