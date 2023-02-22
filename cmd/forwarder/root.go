// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"github.com/saucelabs/forwarder/cmd/forwarder/paceval"
	"github.com/saucelabs/forwarder/cmd/forwarder/pacserver"
	"github.com/saucelabs/forwarder/cmd/forwarder/proxy"
	"github.com/saucelabs/forwarder/cmd/forwarder/version"
	"github.com/spf13/cobra"
)

const (
	envPrefix = "FORWARDER"
	maxCols   = 80
)

func rootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "forwarder",
		Short: "A simple flexible forward proxy",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return bindFlagsToEnv(cmd, envPrefix)
		},
	}
	rootCmd.AddCommand(
		withPACSupportedFunctions(paceval.Command()),
		withPACSupportedFunctions(pacserver.Command()),
		withPACSupportedFunctions(proxy.Command()),
		version.Command(),
	)
	decorateRootCmd(rootCmd)

	for _, cmd := range rootCmd.Commands() {
		appendEnvToUsage(cmd, envPrefix)
		wrapLongAt(cmd, maxCols)
	}

	return rootCmd
}
