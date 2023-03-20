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
	"strings"

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
	flagStr := flagAndUsage[0]

	// if the flag usage is longer than one line, wrap it again
	if len(flagAndUsage) > 1 {
		nextLines := strings.Join(flagAndUsage[1:], " ")
		wrappedUsages := wordwrap.WrapString(nextLines, p.wrapLimit-offset)
		wrappedStr = flagStr + "\n" + wrappedUsages
	}
	appendTabStr := strings.ReplaceAll(wrappedStr, "\n", "\n\t")

	fmt.Fprintf(p.out, appendTabStr+"\n\n")
}

// writeFlag will output the help flag based
// on the format provided by getFlagFormat to i/o writer
func writeFlag(out io.Writer, f *flag.Flag, envPrefix string) {
	val, usage := flag.UnquoteUsage(f)
	if val == "string" {
		val = ""
	}
	if val != "" {
		val = " " + val
	}

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

	fmt.Fprintf(out, getFlagFormat(f), f.Shorthand, f.Name, val, def, env, usage, deprecated)
}

func envName(envPrefix, flagName string) string {
	name := flagName
	name = strings.ReplaceAll(name, "-", "_")
	name = fmt.Sprintf("%s_%s", envPrefix, name)
	return strings.ToUpper(name)
}
