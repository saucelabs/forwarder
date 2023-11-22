// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ReadURLString can read base64 encoded data, local file, http or https URL or stdin and return it as a string.
func ReadURLString(u *url.URL, rt http.RoundTripper) (string, error) {
	b, err := ReadURL(u, rt)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ReadURL can read base64 encoded data, local file, http or https URL or stdin.
func ReadURL(u *url.URL, rt http.RoundTripper) ([]byte, error) {
	switch u.Scheme {
	case "data":
		return readData(u)
	case "file":
		return readFile(u)
	case "http", "https":
		return readHTTP(u, rt)
	default:
		return nil, fmt.Errorf("unsupported scheme %q, supported schemes are: file, http and https", u.Scheme)
	}
}

func readData(u *url.URL) ([]byte, error) {
	v := strings.TrimPrefix(u.Opaque, "//")

	idx := strings.IndexByte(v, ',')
	if idx != -1 {
		if v[:idx] != "base64" {
			return nil, fmt.Errorf("invalid data URI, the only supported format is: data:base64,<encoded data>")
		}
		v = v[idx+1:]
	}

	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func readFile(u *url.URL) ([]byte, error) {
	if u.Host != "" {
		return nil, fmt.Errorf("invalid file URL %q, host is not allowed", u.String())
	}
	if u.User != nil {
		return nil, fmt.Errorf("invalid file URL %q, user is not allowed", u.String())
	}
	if u.RawQuery != "" {
		return nil, fmt.Errorf("invalid file URL %q, query is not allowed", u.String())
	}
	if u.Fragment != "" {
		return nil, fmt.Errorf("invalid file URL %q, fragment is not allowed", u.String())
	}
	if u.Path == "" {
		return nil, fmt.Errorf("invalid file URL %q, path is empty", u.String())
	}

	if u.Path == "-" {
		return readAndClose(os.Stdin)
	}

	f, err := os.Open(u.Path)
	if err != nil {
		return nil, err
	}
	return readAndClose(f)
}

func readAndClose(r io.ReadCloser) ([]byte, error) {
	defer r.Close()
	return io.ReadAll(r)
}

func readHTTP(u *url.URL, rt http.RoundTripper) ([]byte, error) {
	c := http.Client{
		Transport: rt,
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), http.NoBody) //nolint:noctx // timeout is set in the transport
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func ReadFileOrBase64(name string) ([]byte, error) {
	if strings.HasPrefix(name, "data:") {
		return readData(&url.URL{
			Scheme: "data",
			Opaque: name[5:],
		})
	}

	return os.ReadFile(name)
}
