/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package templates

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/mitchellh/go-wordwrap"
	flag "github.com/spf13/pflag"
)

const offset = 10

// HelpFlagPrinter is a printer that
// processes the help flag and print
// it to i/o writer
type HelpFlagPrinter struct {
	envPrefix string
	wrapLimit uint
	out       io.Writer
}

// NewHelpFlagPrinter will initialize a HelpFlagPrinter given the
// i/o writer
func NewHelpFlagPrinter(out io.Writer, envPrefix string, wrapLimit uint) *HelpFlagPrinter {
	return &HelpFlagPrinter{
		envPrefix: envPrefix,
		wrapLimit: wrapLimit,
		out:       out,
	}
}

// PrintHelpFlag will beautify the help flags and print it out to p.out
func (p *HelpFlagPrinter) PrintHelpFlag(flag *flag.Flag) {
	formatBuf := new(bytes.Buffer)
	writeFlag(formatBuf, flag, p.envPrefix)

	wrappedStr := formatBuf.String()
	flagAndUsage := strings.Split(formatBuf.String(), "\n")

	// It always has at least 2 elements.
	flagStr, usage := flagAndUsage[0], strings.Join(flagAndUsage[1:], " ")

	usage = strings.ReplaceAll(usage, "<br>", "\n")
	usage = strings.ReplaceAll(usage, "<ul>", "")
	usage = strings.ReplaceAll(usage, "<li>", "\n- ")
	usage = strings.ReplaceAll(usage, "</ul>", "\n\n")
	usage = strings.ReplaceAll(usage, "<code>", "")
	usage = strings.ReplaceAll(usage, "</code>", "")
	usage = strings.ReplaceAll(usage, "<code-block>", "\"")
	usage = strings.ReplaceAll(usage, "</code-block>", "\"")
	usage = withLinks(usage)
	usage = strings.TrimSpace(usage)

	usage = wordwrap.WrapString(usage, p.wrapLimit-offset)
	wrappedStr = flagStr + "\n" + usage
	appendTabStr := strings.ReplaceAll(wrappedStr, "\n", "\n\t")

	fmt.Fprintf(p.out, appendTabStr+"\n\n")
}

// writeFlag will output the help flag based
// on the format provided by getFlagFormat to i/o writer
func writeFlag(out io.Writer, f *flag.Flag, envPrefix string) {
	name, usage := flagNameAndUsage(f)

	def := f.DefValue
	if def == "[]" {
		def = ""
	}
	if def != "" {
		if f.Value.Type() == "string" {
			def = fmt.Sprintf(" (default '%s')", f.DefValue)
		} else {
			def = fmt.Sprintf(" (default %s)", f.DefValue)
		}
	}

	deprecated := ""
	if f.Deprecated != "" {
		deprecated = fmt.Sprintf(" (DEPRECATED: %s)", f.Deprecated)
	}

	env := fmt.Sprintf(" (env %s)", envName(envPrefix, f.Name))

	fmt.Fprintf(out, getFlagFormat(f), f.Shorthand, f.Name, name, def, env, usage, deprecated)
}

func flagNameAndUsage(f *flag.Flag) (string, string) {
	name, usage := flag.UnquoteUsage(f)

	if vt := findValueType(usage); vt > 0 {
		name = usage[:vt]
		usage = strings.TrimSpace(usage[vt:])
	} else {
		if name == "" || name == "string" {
			name = "value"
		}
		name = fmt.Sprintf("<%s>", name)
	}
	if name != "" {
		name = " " + name
	}

	return name, usage
}

func findValueType(usage string) int {
	runes := []rune(usage)
	if len(runes) == 0 {
		return 0
	}

	var (
		a, b  rune
		stack int
	)
	update := func(r rune) {
		switch r {
		case '<':
			a, b = '<', '>'
			stack = 1
		case '[':
			a, b = '[', ']'
			stack = 1
		}
	}
	update(runes[0])

	if stack == 0 {
		return 0
	}

	for i := 1; i < len(runes); i++ {
		if stack == 0 {
			if unicode.IsUpper(runes[i]) {
				return i
			}
			update(runes[i])
		} else {
			switch runes[i] {
			case a:
				stack++
			case b:
				stack--
			}
		}
	}

	if stack > 0 {
		panic("unbalanced brackets in usage string")
	}

	return len(runes)
}

func withLinks(s string) string {
	re := regexp.MustCompile(linkPattern)

	result := re.ReplaceAllStringFunc(s, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 4 {
			// If the match does not have two groups (text, root, and path), return the match as is.
			return match
		}
		text, root, path := submatches[1], submatches[2], submatches[3]

		if text != "" {
			text = text + ": "
		}

		return text + root + path
	})

	return result
}

var envReplacer = strings.NewReplacer(".", "_", "-", "_")

func envName(envPrefix, flagName string) string {
	s := fmt.Sprintf("%s_%s", envPrefix, flagName)
	s = strings.ToUpper(s)
	s = envReplacer.Replace(s)
	return s
}
