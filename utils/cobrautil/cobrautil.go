// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"github.com/spf13/cobra"
)

// DefaultLong sets the long description to the short description if the long description is empty.
func DefaultLong(cmd *cobra.Command) {
	if cmd.Short == "" {
		return
	}

	if cmd.Long == "" {
		cmd.Long = cmd.Short + "."
	} else {
		cmd.Long = cmd.Short + ".\n\n" + cmd.Long
	}
}

func NoHelpSubcommand(cmd *cobra.Command) {
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
}
