// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
	for range 10000 {
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
