// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// ReadURL can read a local file, http or https URL or stdin.
func ReadURL(u *url.URL, rt http.RoundTripper) (string, error) {
	switch u.Scheme {
	case "file":
		return readFile(u)
	case "http", "https":
		return readHTTP(u, rt)
	default:
		return "", fmt.Errorf("unsupported scheme %q, supported schemes are: file, http and https", u.Scheme)
	}
}

func readFile(u *url.URL) (string, error) {
	if u.Host != "" {
		return "", fmt.Errorf("invalid file URL %q, host is not allowed", u.String())
	}
	if u.User != nil {
		return "", fmt.Errorf("invalid file URL %q, user is not allowed", u.String())
	}
	if u.RawQuery != "" {
		return "", fmt.Errorf("invalid file URL %q, query is not allowed", u.String())
	}
	if u.Fragment != "" {
		return "", fmt.Errorf("invalid file URL %q, fragment is not allowed", u.String())
	}
	if u.Path == "" {
		return "", fmt.Errorf("invalid file URL %q, path is empty", u.String())
	}

	if u.Path == "-" {
		return readAndClose(os.Stdin)
	}

	f, err := os.Open(u.Path)
	if err != nil {
		return "", err
	}
	return readAndClose(f)
}

func readAndClose(r io.ReadCloser) (string, error) {
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readHTTP(u *url.URL, rt http.RoundTripper) (string, error) {
	c := http.Client{
		Transport: rt,
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), http.NoBody) //nolint:noctx // timeout is set in the transport
	if err != nil {
		return "", err
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
