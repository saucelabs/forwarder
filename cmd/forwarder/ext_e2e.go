// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build e2e

package main

import (
	"github.com/saucelabs/forwarder/cmd/forwarder/httpbin"
	"github.com/spf13/cobra"
)

func decorateRootCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(httpbin.Command())
}
