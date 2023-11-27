// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/saucelabs/forwarder/command/forwarder"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
)

var (
	docsDir = flag.String("docs-dir", "", "path to the docs directory")

	cliDir, cfgDir string
)

func main() {
	flag.Parse()

	cliDir = path.Join(*docsDir, "content", "cli")
	cfgDir = path.Join(*docsDir, "content", "config")

	for _, dir := range []string{cliDir, cfgDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			log.Fatal(err)
		}
	}

	cg := forwarder.CommandGroups()
	cg.Add(&cobra.Command{
		Use: "forwarder",
	})
	if err := writeCommandIndex(cg); err != nil {
		log.Fatal(err)
	}

	if err := writeCommandDoc(forwarder.Command()); err != nil {
		log.Fatal(err)
	}

	if err := writeDefaultConfig(forwarder.Command()); err != nil {
		log.Fatal(err)
	}
}

var cmdIndex map[string]int

func writeCommandIndex(cg templates.CommandGroups) error {
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

	fmt.Fprint(f, "# Forwarder CLI\n\n")
	for _, item := range items {
		fmt.Fprintf(f, "%s\n", item)
	}

	return f.Close()
}

const cliWeight = 100

func writeCommandDoc(cmd *cobra.Command) error {
	for _, cmd := range cmd.Commands() {
		if err := writeCommandDoc(cmd); err != nil {
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
		fmt.Fprintf(f, "title: %s\n", cobrautil.FullCommandName(cmd))
		if idx, ok := cmdIndex[cobrautil.FullCommandName(cmd)]; ok {
			fmt.Fprintf(f, "weight: %d\n", cliWeight+idx)
		}
		fmt.Fprintf(f, "---\n\n")

		cobrautil.WriteMarkdownDoc(f, forwarder.FlagGroups(), forwarder.EnvPrefix, cmd)

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

func writeDefaultConfig(cmd *cobra.Command) error {
	for _, cmd := range cmd.Commands() {
		if err := writeDefaultConfig(cmd); err != nil {
			return err
		}
	}

	if cmd.IsAvailableCommand() && cmd.Flags().HasFlags() {
		f, err := os.Create(path.Join(cfgDir, fileName(cmd, ".yaml")))
		if err != nil {
			return err
		}

		cobrautil.WriteConfigFile(f, forwarder.FlagGroups(), forwarder.ConfigFileFlagName, cmd.Flags())

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
