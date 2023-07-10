// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"net"
	"net/http"
	"os"
	"testing"
)

func TestProxyLocalhost(t *testing.T) {
	hosts := []string{
		"localhost",
		"127.0.0.1",
	}

	for _, h := range hosts {
		if os.Getenv("FORWARDER_PROXY_LOCALHOST") == "allow" {
			newClient(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").ExpectStatus(http.StatusOK)
		} else {
			newClient(t, "http://"+net.JoinHostPort(h, "10000")).GET("/version").ExpectStatus(http.StatusBadGateway)
		}
	}
}

func TestSC2450(t *testing.T) {
	if os.Getenv("FORWARDER_SC2450") == "" {
		t.Skip("FORWARDER_SC2450 not set")
	}

	c := newClient(t, "http://sc-2450:8307")
	c.HEAD("/").ExpectStatus(http.StatusOK)
	c.GET("/").ExpectStatus(http.StatusOK).ExpectBodyContent(`{"android":{"min_version":"4.0.0"},"ios":{"min_version":"4.0.0"}}`)
}

func TestHeaderMods(t *testing.T) {
	if os.Getenv("FORWARDER_TEST_HEADERS") == "" {
		t.Skip("FORWARDER_TEST_HEADERS not set")
	}

	c := newClient(t, httpbin)
	c.GET("/header/test-add/test-value").ExpectStatus(http.StatusOK)
	c.GET("/header/test-empty/", func(r *http.Request) {
		r.Header.Set("test-empty", "not-empty")
	}).ExpectStatus(http.StatusOK)
	c.GET("/header/test-rm/value-1", func(r *http.Request) {
		r.Header.Set("test-rm", "value-1")
	}).ExpectStatus(http.StatusNotFound)
	c.GET("/header/rm-prefix/value-2", func(r *http.Request) {
		r.Header.Set("rm-prefix", "value-2")
	}).ExpectStatus(http.StatusNotFound)
}

func TestHeaderRespMods(t *testing.T) {
	if os.Getenv("FORWARDER_TEST_RESPONSE_HEADERS") == "" {
		t.Skip("FORWARDER_TEST_RESPONSE_HEADERS not set")
	}

	c := newClient(t, httpbin)
	c.GET("/status/200").ExpectStatus(http.StatusOK).ExpectHeader("test-resp-add", "test-resp-value")
	c.GET("/header/test-resp-empty/not-empty", func(r *http.Request) {
		r.Header.Set("test-resp-empty", "not-empty")
	}).ExpectStatus(http.StatusOK).ExpectHeader("test-resp-empty", "")
	c.GET("/header/test-resp-rm/value-3", func(r *http.Request) {
		r.Header.Set("test-resp-rm", "value-3")
	}).ExpectStatus(http.StatusOK).ExpectHeader("test-resp-rm", "")
	c.GET("/header/resp-rm-prefix/value-4", func(r *http.Request) {
		r.Header.Set("resp-rm-prefix", "value-4")
	}).ExpectStatus(http.StatusOK).ExpectHeader("resp-rm-prefix", "")
}
