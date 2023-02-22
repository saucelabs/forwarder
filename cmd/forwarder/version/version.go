// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package version

import (
	"fmt"
	"runtime"

	"github.com/saucelabs/forwarder/internal/version"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			w := cmd.OutOrStdout()

			fmt.Fprintln(w, "Version:\t", version.Version)
			fmt.Fprintln(w, "Built time:\t", version.Time)
			fmt.Fprintln(w, "Git commit:\t", version.Commit)

			fmt.Fprintln(w, "Go Arch:\t", runtime.GOARCH)
			fmt.Fprintln(w, "Go OS:\t\t", runtime.GOOS)
			fmt.Fprintln(w, "Go Version:\t", runtime.Version())
		},
	}
}
