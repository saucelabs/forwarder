// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"net/http"
	"time"

	"github.com/google/go-github/v56/github"
)

const (
	owner = "saucelabs"
	repo  = "forwarder"
)

func newGitHubClient() *github.Client {
	const timeout = 30 * time.Second

	return github.NewClient(&http.Client{
		Timeout: timeout,
	})
}
