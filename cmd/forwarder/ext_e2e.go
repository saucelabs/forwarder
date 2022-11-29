//go:build e2e

package main

import (
	"github.com/saucelabs/forwarder/cmd/forwarder/httpbin"
	"github.com/spf13/cobra"
)

func decorateRootCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(httpbin.Command())
}
