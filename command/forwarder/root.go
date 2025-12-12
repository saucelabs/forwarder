// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"github.com/saucelabs/forwarder/bind"
	"github.com/saucelabs/forwarder/command/pac"
	"github.com/saucelabs/forwarder/command/ready"
	"github.com/saucelabs/forwarder/command/run"
	"github.com/saucelabs/forwarder/command/test/grpc"
	"github.com/saucelabs/forwarder/command/test/httpbin"
	"github.com/saucelabs/forwarder/command/version"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
)

const (
	EnvPrefix          = "FORWARDER"
	ConfigFileFlagName = "config-file"
)

func CommandGroups() templates.CommandGroups {
	return templates.CommandGroups{
		{
			Message: "Commands:",
			Commands: []*cobra.Command{
				run.Command(),
				pac.Command(),
				ready.Command(),
			},
		},
	}
}

func FlagGroups() templates.FlagGroups {
	return templates.FlagGroups{
		{
			Name: "Server options",
			Prefix: []string{
				"proxy-protocol",
				"",
			},
		},
		{
			Name: "Proxy options",
			Prefix: []string{
				"proxy",
				"pac",

				"direct-domains",
				"deny-domains",

				"header",
				"connect-header",
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
			Name:   "Kerberos options",
			Prefix: []string{"kerberos"},
		},
		{
			Name: "HTTP client options",
			Prefix: []string{
				"http",
				"cacert-file",
				"connect-to",
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
}

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forwarder",
		Short: "HTTP (forward) proxy server with PAC support and PAC testing tools",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cobrautil.BindAll(cmd, EnvPrefix, ConfigFileFlagName)
		},
	}
	bind.ConfigFile(cmd.PersistentFlags(), new(string))

	cg := CommandGroups()
	cg.Add(cmd)

	templates.ActsAsRootCommand(cmd, nil, cg, FlagGroups(), EnvPrefix)

	// Add test commands.
	test := &cobra.Command{
		Use:   "test",
		Short: "Run test servers for various protocols",
	}
	test.AddCommand(
		grpc.Command(),
		httpbin.Command(),
	)
	cmd.AddCommand(test)

	// Add version command.
	cmd.AddCommand(version.Command())

	// Add config-file command to all commands.
	cobrautil.AddConfigFileForEachCommand(cmd, FlagGroups(), ConfigFileFlagName)

	return cmd
}
