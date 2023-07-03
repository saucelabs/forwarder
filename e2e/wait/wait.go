// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package wait

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Waiter struct {
	Client http.Client
	// Endpoint is the endpoint to check for readiness.
	Endpoint string
	// MaxWait is the maximum time to wait for a server to be ready.
	MaxWait time.Duration
	// Backoff is the time to wait between checks.
	Backoff time.Duration
}

var defaultWaiter = Waiter{
	MaxWait:  30 * time.Second,
	Backoff:  500 * time.Millisecond, //nolint:gomnd // default value
	Endpoint: "/readyz",
}

func (w *Waiter) WaitForServerReady(baseURL string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, u.String()+w.Endpoint, http.NoBody)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(w.MaxWait)
	for {
		resp, err := w.Client.Do(req.Clone(context.Background()))
		if resp != nil {
			resp.Body.Close() //noline:errcheck // we don't care about the body

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("not ready after %s: %w", w.MaxWait, err)
			}

			return fmt.Errorf("not ready after %s: status code %d", w.MaxWait, resp.StatusCode)
		}

		time.Sleep(w.Backoff)
	}
}

// ServerReady checks the API server /readyz endpoint until it returns 200.
// It returns an error if the server is not ready after 30 seconds.
// See Waiter.WaitForServerReady for more details.
func ServerReady(baseURL string) error {
	return defaultWaiter.WaitForServerReady(baseURL)
}
