// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

//go:build e2e

package main

import (
	"github.com/saucelabs/forwarder/cmd/forwarder/httpbin"
	"github.com/spf13/cobra"
)

func decorateRootCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(httpbin.Command())
}
