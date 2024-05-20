// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package templates

import (
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
	name, usage := flagNameAndUsage(f)

	deprecated := ""
	if f.Deprecated != "" {
		deprecated = fmt.Sprintf("\nDEPRECATED: %s", f.Deprecated)
	}

	usage = strings.ReplaceAll(usage, "<br>", "\n")
	usage = strings.ReplaceAll(usage, "<ul>", "")
	usage = strings.ReplaceAll(usage, "<li>", "\n- ")
	usage = strings.ReplaceAll(usage, "</ul>", "\n\n")
	usage = strings.ReplaceAll(usage, "<code>", "\"")
	usage = strings.ReplaceAll(usage, "</code>", "\"")
	usage = withLinks(usage)

	fmt.Fprintf(p.out, "# %s%s\n#\n", f.Name, name)
	for _, l := range strings.Split(wordwrap.WrapString(usage, p.wrapLimit-2), "\n") {
		fmt.Fprintf(p.out, "# %s\n", l)
	}
	if deprecated != "" {
		fmt.Fprintf(p.out, "# %s\n", deprecated)
	}
	fmt.Fprintf(p.out, "#%s: %s\n\n", f.Name, p.defaultValue(f))
}

func (p *YamlFlagPrinter) defaultValue(f *pflag.Flag) string {
	def := f.DefValue
	if def == "[]" {
		def = ""
	}
	return def
}
