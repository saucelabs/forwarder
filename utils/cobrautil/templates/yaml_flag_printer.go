// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package templates

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/pflag"
)

type YamlFlagPrinter struct {
	out       io.Writer
	wrapLimit uint
}

func NewYamlFlagPrinter(out io.Writer, wrapLimit uint) *YamlFlagPrinter {
	return &YamlFlagPrinter{
		out:       out,
		wrapLimit: wrapLimit,
	}
}

func (p *YamlFlagPrinter) PrintHelpFlag(f *pflag.Flag) {
	formatBuf := new(bytes.Buffer)
	writeYamlFlag(formatBuf, f)

	wrappedStr := formatBuf.String()
	flagAndUsage := strings.Split(formatBuf.String(), "\n")

	// if the flag usage is longer than one line, wrap it again
	if len(flagAndUsage) > 1 {
		nextLines := strings.Join(flagAndUsage[:len(flagAndUsage)-1], " ")
		wrappedUsages := wordwrap.WrapString(nextLines, p.wrapLimit-2)
		wrappedUsages = "#\n# " + strings.ReplaceAll(wrappedUsages, "\n", "\n# ")
		wrappedStr = wrappedUsages + "\n#\n#" + flagAndUsage[len(flagAndUsage)-1]
	}
	fmt.Fprintf(p.out, wrappedStr)
	fmt.Fprintf(p.out, "\n\n")
}

func writeYamlFlag(out io.Writer, f *pflag.Flag) {
	_, usage := flagNameAndUsage(f)

	def := f.DefValue
	if def == "[]" {
		def = ""
	}
	if def != "" {
		def = " " + def
	}

	deprecated := ""
	if f.Deprecated != "" {
		deprecated = fmt.Sprintf("\nDEPRECATED: %s", f.Deprecated)
	}

	fmt.Fprintf(out, "%s%s\n%s:%s", usage, deprecated, f.Name, def)
}
