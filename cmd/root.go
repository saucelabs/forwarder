// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/saucelabs/sypl"
	"github.com/saucelabs/sypl/level"
	"github.com/spf13/cobra"
)

var (
	cliLogger *sypl.Sypl

	logLevel, fileLevel, filePath string
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "proxy",
	Short: "A simple flexible, and proxy",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() { _ = rootCmd.Execute() }

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "sets the log level (default info)")
	rootCmd.PersistentFlags().StringVar(&fileLevel, "log-file-level", "info", "sets the log file level (default info)")
	rootCmd.PersistentFlags().StringVar(&filePath, "log-file-path", "", `sets the log file path (default "OS temp dir")`)

	cliLogger = sypl.NewDefault("proxy", level.MustFromString(logLevel))
}
