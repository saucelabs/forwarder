// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"github.com/saucelabs/forwarder/cmd/forwarder/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "proxy",
	Short: "A simple flexible forward proxy",
}

func init() {
	rootCmd.AddCommand(
		version.Command(),
	)
}
