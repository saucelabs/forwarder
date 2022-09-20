// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	buildVersion = "Devel"
	buildTime    = "Unknown"
	buildCommit  = "Unknown"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints version information",
		Run: func(cmd *cobra.Command, args []string) {
			w := cmd.OutOrStdout()

			fmt.Fprintln(w, "Version:\t", buildVersion)
			fmt.Fprintln(w, "Built time:\t", buildTime)
			fmt.Fprintln(w, "Git commit:\t", buildCommit)

			fmt.Fprintln(w, "Go Arch:\t", runtime.GOARCH)
			fmt.Fprintln(w, "Go OS:\t\t", runtime.GOOS)
			fmt.Fprintln(w, "Go Version:\t", runtime.Version())
		},
	}
}
