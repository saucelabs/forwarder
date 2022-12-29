// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package main

import (
	"regexp"
	"strings"

	"github.com/saucelabs/forwarder/pac"
	"github.com/spf13/cobra"
)

func withPACSupportedFunctions(cmd *cobra.Command) *cobra.Command {
	cmd.Example += "\n" + pacSupportedFunctions()
	return cmd
}

func pacSupportedFunctions() string {
	var sb strings.Builder
	sb.WriteString("Supported PAC util functions:")
	for _, fn := range pac.SupportedFunctions() {
		sb.WriteString("\n  ")
		sb.WriteString(fn)
	}
	return sb.String()
}

func wrapLongAt(cmd *cobra.Command, width int) {
	cmd.Long = wrapTextAt(cmd.Long, width)
}

func wrapTextAt(s string, width int) string {
	s = regexp.MustCompile(`\n{2,}`).ReplaceAllString(s, "\n<separator>\n")

	var (
		sb strings.Builder
		lw = 0
	)
	for _, w := range regexp.MustCompile(`\s+`).Split(s, -1) {
		if w == "<separator>" {
			sb.WriteString("\n\n")
			lw = 0
			continue
		}

		if lw+len(w) > width {
			sb.Write([]byte{'\n'})
			lw = 0
		}
		sb.WriteString(w)
		sb.Write([]byte{' '})
		lw += len(w) + 1
	}

	return "\n" + sb.String()
}
