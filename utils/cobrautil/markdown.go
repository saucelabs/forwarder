// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"fmt"
	"io"

	"github.com/saucelabs/forwarder/utils/cobrautil/templates"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func WriteMarkdownDoc(w io.Writer, g templates.FlagGroups, envPrefix string, cmd *cobra.Command) {
	p := templates.NewMarkdownFlagPrinter(w, envPrefix)

	fmt.Fprintf(w, "# %s\n\n", FullCommandNameTitle(cmd))

	fmt.Fprintf(w, "Usage: `%s`\n\n", cmd.UseLine())

	if cmd.Long != "" {
		fmt.Fprintf(w, "%s\n\n", cmd.Long)
	} else if cmd.Short != "" {
		fmt.Fprintf(w, "%s\n\n", cmd.Short)
	}

	if cmd.Example != "" {
		fmt.Fprintf(w, "## Examples\n\n```\n%s\n```\n\n", cmd.Example)
	}

	for i, fs := range templates.SplitFlagSet(g, cmd.Flags()) {
		if !fs.HasAvailableFlags() {
			continue
		}

		fmt.Fprintf(w, "## %s\n\n", g[i].Name)

		fs.VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden {
				return
			}

			p.PrintHelpFlag(flag)
		})
	}
}
