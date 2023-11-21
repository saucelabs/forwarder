// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:generate go run gen.go

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/saucelabs/forwarder/cmd/forwarder/root"
	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/spf13/cobra"
)

func main() {
	cmd := root.Command()

	if err := writeCommandDoc(cmd); err != nil {
		log.Fatal(err)
	}

	if err := writeIndex(cmd); err != nil {
		log.Fatal(err)
	}
}

func writeCommandDoc(cmd *cobra.Command) error {
	for _, cmd := range cmd.Commands() {
		if err := writeCommandDoc(cmd); err != nil {
			return err
		}
	}

	if cmd.IsAvailableCommand() && cmd.Flags().HasFlags() {
		f, err := os.Create(mdFileName(cmd))
		if err != nil {
			return err
		}

		cobrautil.WriteMarkdownDoc(f, root.FlagGroups(), root.EnvPrefix, cmd)

		return f.Close()
	}

	return nil
}

func mdFileName(cmd *cobra.Command) string {
	return strings.ReplaceAll(cobrautil.FullCommandName(cmd), " ", "_") + ".md"
}

func writeIndex(cmd *cobra.Command) error {
	fileName := "index.md"

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	var items []string
	walkCommand(cmd, func(cmd *cobra.Command) {
		if cmd.IsAvailableCommand() && cmd.Flags().HasFlags() {
			items = append(items, fmt.Sprintf("- [%s](%s) - %s", cobrautil.FullCommandName(cmd), mdFileName(cmd), cmd.Short))
		}
	})

	fmt.Fprintf(f, "# %s\n\n", cmd.Name())
	for _, item := range items {
		fmt.Fprintf(f, "%s\n", item)
	}

	return f.Close()
}

func walkCommand(cmd *cobra.Command, f func(cmd *cobra.Command)) {
	f(cmd)
	for _, cmd := range cmd.Commands() {
		walkCommand(cmd, f)
	}
}
