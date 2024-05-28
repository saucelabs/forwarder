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

	usage = p.replaceHTML(usage)
	usage = withLinks(usage)

	fmt.Fprintf(p.out, "# %s%s\n#\n", f.Name, name)
	for _, l := range strings.Split(wordwrap.WrapString(usage, p.wrapLimit-2), "\n") {
		fmt.Fprintf(p.out, "# %s\n", l)
	}
	if f.Deprecated != "" {
		fmt.Fprintf(p.out, "#\n# DEPRECATED: %s\n", f.Deprecated)
	}
	fmt.Fprintf(p.out, "#%s: %s\n\n", f.Name, p.defaultValue(f))
}

func (p *YamlFlagPrinter) replaceHTML(s string) string {
	r := strings.NewReplacer(
		"<br>", "\n",
		"<ul>", "",
		"<li>", "\n- ",
		"</ul>", "\n\n",
		"<code>", "",
		"</code>", "",
		"<code-block>", "\n\n",
		"</code-block>", "\n\n",
	)

	s = r.Replace(s)
	s = strings.TrimSpace(s)
	return s
}

func (p *YamlFlagPrinter) defaultValue(f *pflag.Flag) string {
	def := f.DefValue
	if def == "[]" {
		def = ""
	}
	return def
}
