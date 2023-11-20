// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"fmt"

	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func ConfigFileCommand(g templates.FlagGroups, fs *pflag.FlagSet, configFileFlagName string) *cobra.Command {
	return &cobra.Command{
		Use:    "config-file",
		Args:   cobra.NoArgs,
		Hidden: true,
		Run: func(cmd *cobra.Command, _ []string) {
			w := cmd.OutOrStdout()
			p := templates.NewYamlFlagPrinter(w, 80)

			for i, fs := range templates.SplitFlagSet(g, fs) {
				if !fs.HasAvailableFlags() {
					continue
				}

				header := true
				fs.VisitAll(func(flag *pflag.Flag) {
					if flag.Hidden {
						return
					}
					if flag.Name == configFileFlagName {
						return
					}

					if header {
						fmt.Fprintf(w, "# --- %s ---\n\n", g[i].Name)
						header = false
					}

					p.PrintHelpFlag(flag)
				})
			}
		},
	}
}

func AddConfigFileForEachCommand(cmd *cobra.Command, g templates.FlagGroups, configFileFlagName string) {
	for _, cmd := range cmd.Commands() {
		AddConfigFileForEachCommand(cmd, g, configFileFlagName)
	}

	if cmd.IsAvailableCommand() && cmd.Flags().HasFlags() {
		cmd.AddCommand(ConfigFileCommand(g, cmd.Flags(), configFileFlagName))
	}
}
