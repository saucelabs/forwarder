// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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
