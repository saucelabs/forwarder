// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/cmd/forwarder/httpbin"
	"github.com/saucelabs/forwarder/cmd/forwarder/pac"
	"github.com/saucelabs/forwarder/cmd/forwarder/ready"
	"github.com/saucelabs/forwarder/cmd/forwarder/run"
	"github.com/saucelabs/forwarder/cmd/forwarder/version"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
)

const (
	EnvPrefix          = "FORWARDER"
	ConfigFileFlagName = "config-file"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forwarder",
		Short: "HTTP (forward) proxy server with PAC support and PAC testing tools",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cobrautil.BindAll(cmd, EnvPrefix, ConfigFileFlagName)
		},
	}
	bind.ConfigFile(cmd.PersistentFlags(), new(string))

	commandGroups := templates.CommandGroups{
		{
			Message: "Commands:",
			Commands: []*cobra.Command{
				run.Command(),
				pac.Command(),
				ready.Command(),
			},
		},
	}
	commandGroups.Add(cmd)

	flagGroups := templates.FlagGroups{
		{
			Name:   "Server options",
			Prefix: []string{""},
		},
		{
			Name: "Proxy options",
			Prefix: []string{
				"proxy",
				"pac",

				"direct-domains",
				"deny-domains",

				"header",
				"proxy-header",
				"response-header",
			},
		},
		{
			Name:   "MITM options",
			Prefix: []string{"mitm"},
		},
		{
			Name:   "DNS options",
			Prefix: []string{"dns"},
		},
		{
			Name: "HTTP client options",
			Prefix: []string{
				"http",
				"cacert-file",
				"insecure",
			},
		},
		{
			Name: "API server options",
			Prefix: []string{
				"api",
				"prom",
			},
		},
		{
			Name:   "Logging options",
			Prefix: []string{"log"},
		},
		{
			Name:   "Options",
			Prefix: []string{"config-file"},
		},
	}

	templates.ActsAsRootCommand(cmd, nil, commandGroups, flagGroups, EnvPrefix)

	// Add other commands.
	cmd.AddCommand(
		httpbin.Command(), // hidden
		version.Command(),
	)

	// Add config-file command to all commands.
	cobrautil.AddConfigFileForEachCommand(cmd, flagGroups, ConfigFileFlagName)

	return cmd
}
