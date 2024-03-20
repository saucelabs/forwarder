// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package docsgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/saucelabs/forwarder/utils/cobrautil"
	"github.com/saucelabs/forwarder/utils/promutil"
	"github.com/spf13/cobra"
)

func WriteCommandProm(cmd *cobra.Command, promDir string) error {
	f, err := os.Create(path.Join(promDir, "metrics.md"))
	if err != nil {
		return err
	}

	fmt.Fprintf(f, "---\n")
	fmt.Fprintf(f, "id: metrics\n")
	fmt.Fprintf(f, "title: Metrics\n")
	fmt.Fprintf(f, "---\n\n")

	fmt.Fprintf(f, "# Prometheus Metrics\n\n")

	for _, cmd := range cmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.Flags().Lookup("desc-metrics") == nil {
			continue
		}
		// We need to copy the command because Execute() calls the root command which messes up the flags.
		c := cobra.Command{
			Run:  cmd.Run,
			RunE: cmd.RunE,
		}
		c.Flags().AddFlagSet(cmd.Flags())

		var buf bytes.Buffer
		c.SetOut(io.Discard)
		c.SetErr(&buf)
		c.SetArgs([]string{"--desc-metrics"})
		if err := c.Execute(); err != nil {
			return err
		}

		var desc []promutil.Desc
		if err := json.NewDecoder(&buf).Decode(&desc); err != nil {
			return err
		}

		fmt.Fprintf(f, "## %s\n", cobrautil.FullCommandName(cmd))
		writePromMarkdown(f, desc)
	}

	return f.Close()
}

func writePromMarkdown(f io.Writer, desc []promutil.Desc) {
	slices.SortFunc(desc, func(a, b promutil.Desc) int {
		ap := a.FqName[:strings.Index(a.FqName, "_")] //nolint:gocritic // _ is guaranteed to be in the string
		bp := b.FqName[:strings.Index(b.FqName, "_")] //nolint:gocritic // _ is guaranteed to be in the string

		if ap == "go" {
			ap = "zz"
		}
		if bp == "go" {
			bp = "zz"
		}
		if c := strings.Compare(ap, bp); c != 0 {
			return c
		}

		return strings.Compare(a.FqName, b.FqName)
	})

	for _, d := range desc {
		fmt.Fprintf(f, "\n### `%s`\n\n%s\n", d.FqName, d.Help)

		if len(d.ConstLabels)+len(d.VariableLabels) > 0 {
			fmt.Fprintf(f, "\nLabels:\n")
		}
		for k := range d.ConstLabels {
			fmt.Fprintf(f, "  - %s\n", k)
		}
		for _, k := range d.VariableLabels {
			fmt.Fprintf(f, "  - %s\n", k)
		}
	}

	fmt.Fprintf(f, "\n")
}
