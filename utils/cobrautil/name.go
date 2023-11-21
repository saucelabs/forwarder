// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// FullCommandName returns the full name of the command from the root command.
func FullCommandName(cmd *cobra.Command) string {
	if cmd.Parent() == nil {
		return cmd.Name()
	}
	return FullCommandName(cmd.Parent()) + " " + cmd.Name()
}

var titleCase = cases.Title(language.English)

// FullCommandNameTitle returns the full name of the command from the root command
// with the first letter of each word capitalized.
func FullCommandNameTitle(cmd *cobra.Command) string {
	return titleCase.String(FullCommandName(cmd))
}
