// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import (
	"github.com/saucelabs/forwarder/command/pac/eval"
	"github.com/saucelabs/forwarder/command/pac/server"
	"github.com/spf13/cobra"
)

func Command() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "pac",
		Short: "Tools for working with PAC files",
	}
	cmd.AddCommand(
		eval.Command(),
		server.Command(),
	)
	return cmd
}
