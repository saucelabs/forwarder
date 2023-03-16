// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"github.com/saucelabs/forwarder/cmd/forwarder/httpbin"
	"github.com/saucelabs/forwarder/cmd/forwarder/pac"
	"github.com/saucelabs/forwarder/cmd/forwarder/run"
	"github.com/saucelabs/forwarder/cmd/forwarder/version"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/spf13/cobra"
)

const (
	envPrefix = "FORWARDER"
	maxCols   = 80
)

func rootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "forwarder",
		Short: "HTTP (forward) proxy server with PAC support and PAC testing tools",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cobrautil.BindFlagsToEnv(cmd, envPrefix)
		},
	}

	rootCmd.AddCommand(
		httpbin.Command(),
		pac.Command(),
		run.Command(),
		version.Command(),
	)
	applyDefaults(rootCmd)

	return rootCmd
}

func applyDefaults(cmd *cobra.Command) {
	cobrautil.AppendEnvToUsage(cmd, envPrefix)
	cobrautil.DefaultLong(cmd)
	cobrautil.NoHelpSubcommand(cmd)
	cobrautil.WrapLong(cmd, maxCols)

	for _, cmd := range cmd.Commands() {
		applyDefaults(cmd)
	}
}
