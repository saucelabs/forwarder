package main

import (
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

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
