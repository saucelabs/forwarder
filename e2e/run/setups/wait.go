// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package setups

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WaitForServerReady checks the API server /readyz endpoint until it returns 200.
func WaitForServerReady(baseURL string) error {
	var client http.Client

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	readyz := fmt.Sprintf("%s/readyz", u)

	req, err := http.NewRequest(http.MethodGet, readyz, http.NoBody)
	if err != nil {
		return err
	}

	const backoff = 200 * time.Millisecond
	const maxWait = 5 * time.Second
	var (
		resp *http.Response
		rerr error
	)
	for i := 0; i < int(maxWait/backoff); i++ {
		resp, rerr = client.Do(req.Clone(context.Background()))

		if resp != nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close() //noline:errcheck // we don't care about the body
			return nil
		}

		time.Sleep(backoff)
	}
	if rerr != nil {
		return fmt.Errorf("%s not ready: %w", u.Hostname(), rerr)
	}

	return fmt.Errorf("%s not ready", u.Hostname())
}
