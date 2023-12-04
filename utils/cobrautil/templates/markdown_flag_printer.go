// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package templates

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/pflag"
)

type MarkdownFlagPrinter struct {
	out       io.Writer
	envPrefix string
}

func NewMarkdownFlagPrinter(out io.Writer, envPrefix string) *MarkdownFlagPrinter {
	return &MarkdownFlagPrinter{
		out:       out,
		envPrefix: envPrefix,
	}
}

func (p *MarkdownFlagPrinter) PrintHelpFlag(f *pflag.Flag) {
	fmt.Fprintf(p.out, p.header(f))
	fmt.Fprint(p.out, "\n\n")

	body := p.body(f)
	body = strings.ReplaceAll(body, ". ", ".\n")
	fmt.Fprintf(p.out, body)
	fmt.Fprintf(p.out, "\n\n")
}

func (p *MarkdownFlagPrinter) header(f *pflag.Flag) string {
	format := "--%s"
	if f.Shorthand != "" {
		format = "-%s, " + format
	} else {
		format = "%s" + format
	}
	format = "### `" + format + "` {#%s}"

	return fmt.Sprintf(format, f.Shorthand, f.Name, f.Name)
}

func (p *MarkdownFlagPrinter) body(f *pflag.Flag) string {
	env := fmt.Sprintf("Environment variable: `%s`", envName(p.envPrefix, f.Name))

	_, usage := flagNameAndUsage(f)

	deprecated := ""
	if f.Deprecated != "" {
		deprecated = fmt.Sprintf("\nDEPRECATED: %s", f.Deprecated)
	}

	def := f.DefValue
	if def == "[]" {
		def = ""
	}
	if def != "" {
		def = "\n\nDefault value: `" + def + "`"
	}

	return fmt.Sprintf("%s\n\n%s%s%s", env, usage, deprecated, def)
}
