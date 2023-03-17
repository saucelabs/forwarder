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
	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
)

const envPrefix = "FORWARDER"

func rootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forwarder",
		Short: "HTTP (forward) proxy server with PAC support and PAC testing tools",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cobrautil.BindFlagsToEnv(cmd, envPrefix)
		},
	}

	commandGroups := templates.CommandGroups{
		{
			Message: "Commands:",
			Commands: []*cobra.Command{
				run.Command(),
				pac.Command(),
			},
		},
	}
	commandGroups.Add(cmd)
	templates.ActsAsRootCommand(cmd, nil, commandGroups...)

	// Add other commands
	cmd.AddCommand(
		httpbin.Command(), // hidden
		version.Command(),
	)

	decorate(cmd)

	return cmd
}

func decorate(cmd *cobra.Command) {
	cobrautil.AppendEnvToUsage(cmd, envPrefix)

	for _, cmd := range cmd.Commands() {
		decorate(cmd)
	}
}
