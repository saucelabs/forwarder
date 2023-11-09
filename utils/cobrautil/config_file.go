// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func ConfigFileCommand(g templates.FlagGroups, fs *pflag.FlagSet) *cobra.Command {
	return &cobra.Command{
		Use:    "config-file",
		Args:   cobra.NoArgs,
		Hidden: true,
		Run: func(cmd *cobra.Command, _ []string) {
			p := templates.NewYamlFlagPrinter(cmd.OutOrStdout(), 80)

			for _, fs := range templates.SplitFlagSet(g, fs) {
				if !fs.HasAvailableFlags() {
					continue
				}

				fs.VisitAll(func(flag *pflag.Flag) {
					if flag.Hidden {
						return
					}
					p.PrintHelpFlag(flag)
				})
			}
		},
	}
}
