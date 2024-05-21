// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package templates

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
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
	body = p.replaceHTML(body)
	body = withMarkdownLinks(body)

	fmt.Fprintf(p.out, body)
	fmt.Fprintf(p.out, "\n\n")
}

func (p *MarkdownFlagPrinter) replaceHTML(s string) string {
	r := strings.NewReplacer(
		". ", ".\n",
		"<br>", "\n",
		"<ul>", "\n",
		"<li>", "\n- ",
		"</ul>", "\n\n",
		"<code>", "`",
		"</code>", "`",
		"<code-block>", "\n```\n",
		"</code-block>", "\n```\n",
	)

	s = r.Replace(s)
	s = strings.TrimSpace(s)
	return s
}

func withMarkdownLinks(s string) string {
	re := regexp.MustCompile(linkPattern)

	result := re.ReplaceAllStringFunc(s, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 4 {
			// If the match does not have two groups (text, root, and path), return the match as is.
			return match
		}
		text, path := submatches[1], submatches[3]

		return fmt.Sprintf("[%s](%s)", text, path)
	})

	return result
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
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "* Environment variable: `%s`\n", envName(p.envPrefix, f.Name))
	format, usage := flagNameAndUsage(f)
	fmt.Fprintf(buf, "* Value Format: `%s`\n", strings.TrimSpace(format))
	def := f.DefValue
	if def == "[]" {
		def = ""
	}
	if def != "" {
		fmt.Fprintf(buf, "* Default value: `%s`\n", def)
	}
	fmt.Fprintln(buf)

	if f.Deprecated != "" {
		fmt.Fprintf(buf, "DEPRECATED: %s\n\n", f.Deprecated)
	}

	fmt.Fprintf(buf, "%s", usage)

	return buf.String()
}
