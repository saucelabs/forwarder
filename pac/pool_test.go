// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package pac

import (
	"net/url"
	"sync"
	"testing"

	"go.uber.org/goleak"
)

func TestProxyResolverPoolHammering(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	defaultQueryURL, err := url.ParseRequestURI("https://www.google.com/")
	if err != nil {
		t.Fatal(err)
	}

	const direct = `function FindProxyForURL(url, host) {
  return "DIRECT";
}
`

	defer goleak.VerifyNone(t)
	pool, err := NewProxyResolverPool(&ProxyResolverConfig{Script: direct}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			if _, err := pool.FindProxyForURL(defaultQueryURL, ""); err != nil {
				panic(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
