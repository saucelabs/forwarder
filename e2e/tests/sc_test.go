// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package tests

import (
	"net/http"
	"testing"
)

func TestSC2450(t *testing.T) {
	c := newClient(t, "http://sc-2450:8307")
	c.HEAD("/").ExpectStatus(http.StatusOK)
	c.GET("/").ExpectStatus(http.StatusOK).ExpectBodyContent(`{"android":{"min_version":"4.0.0"},"ios":{"min_version":"4.0.0"}}`)
}
