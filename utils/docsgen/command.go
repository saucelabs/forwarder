// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package docsgen

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/saucelabs/forwarder/command/forwarder"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
)

var (
	FlagGroups = forwarder.FlagGroups()
	EnvPrefix  = forwarder.EnvPrefix

	cmdIndex map[string]int
)

func WriteCommandIndex(cg templates.CommandGroups, cliDir, title string) error {
	cmdIndex = make(map[string]int)

	f, err := os.Create(path.Join(cliDir, "toc.md"))
	if err != nil {
		return err
	}

	var items []string
	for _, g := range cg {
		for _, cmd := range g.Commands {
			walkCommand(cmd, func(cmd *cobra.Command) {
				if cmd.IsAvailableCommand() && cmd.Flags().HasFlags() {
					items = append(items, fmt.Sprintf("- [%s](%s) - %s",
						cobrautil.FullCommandName(cmd),
						fileName(cmd, ".md"),
						cmd.Short),
					)
					cmdIndex[cobrautil.FullCommandName(cmd)] = len(items)
				}
			})
		}
	}

	fmt.Fprintf(f, "---\nbookHidden: true\n---\n\n")

	fmt.Fprintf(f, "# %s CLI\n\n", title)
	for _, item := range items {
		fmt.Fprintf(f, "%s\n", item)
	}

	return f.Close()
}

const cliWeight = 100

func WriteCommandDoc(cmd *cobra.Command, cliDir string) error {
	for _, cmd := range cmd.Commands() {
		if err := WriteCommandDoc(cmd, cliDir); err != nil {
			return err
		}
	}

	if cmd.IsAvailableCommand() && cmd.Flags().HasFlags() {
		f, err := os.Create(path.Join(cliDir, fileName(cmd, ".md")))
		if err != nil {
			return err
		}

		if cmd.Long != "" {
			cmd.Long += "\n\n" + configFileNote(cmd)
		} else if cmd.Short != "" {
			cmd.Short += "\n\n" + configFileNote(cmd)
		}

		fmt.Fprintf(f, "---\n")
		fmt.Fprintf(f, "id: %s\n", cmd.Name())
		fmt.Fprintf(f, "title: %s\n", cobrautil.FullCommandName(cmd))
		if idx, ok := cmdIndex[cobrautil.FullCommandName(cmd)]; ok {
			fmt.Fprintf(f, "weight: %d\n", cliWeight+idx)
		}
		fmt.Fprintf(f, "---\n\n")

		cobrautil.WriteMarkdownDoc(f, FlagGroups, EnvPrefix, cmd)

		return f.Close()
	}

	return nil
}

func configFileNote(cmd *cobra.Command) string {
	return fmt.Sprintf(
		"**Note:** You can also specify the options as YAML, JSON or TOML file using `--config-file` flag.\n"+
			"You can generate a config file by running `%s` command.\n",
		cobrautil.FullCommandName(cmd)+" config-file",
	)
}

func WriteDefaultConfig(cmd *cobra.Command, cfgDir string) error {
	for _, cmd := range cmd.Commands() {
		if err := WriteDefaultConfig(cmd, cfgDir); err != nil {
			return err
		}
	}

	if cmd.IsAvailableCommand() && cmd.Flags().HasFlags() {
		f, err := os.Create(path.Join(cfgDir, fileName(cmd, ".yaml")))
		if err != nil {
			return err
		}

		cobrautil.WriteConfigFile(f, FlagGroups, forwarder.ConfigFileFlagName, cmd.Flags())

		return f.Close()
	}

	return nil
}

func fileName(cmd *cobra.Command, ext string) string {
	return strings.ReplaceAll(cobrautil.FullCommandName(cmd), " ", "_") + ext
}

func walkCommand(cmd *cobra.Command, f func(cmd *cobra.Command)) {
	f(cmd)
	for _, cmd := range cmd.Commands() {
		walkCommand(cmd, f)
	}
}
