/*
Copyright 2016 The Kubernetes Authors.

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
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/saucelabs/forwarder/utils/cobrautil/term"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var DefaultWrapLimit uint = 120

type FlagExposer interface {
	ExposeFlags(cmd *cobra.Command, flags ...string) FlagExposer
}

// HelpOptions enables optional parts of the root command help.
// The zero value keeps the default help output.
type HelpOptions struct {
	// RootUsage shows "Usage: <root> <command> [options]" for a root command that only has subcommands.
	RootUsage bool
	// PersistentFlagsHeading replaces the default persistent flags heading, e.g. "Global options:".
	// When set, the persistent flags and the help flag are listed without flag group names.
	PersistentFlagsHeading string
	// UsageSections sets the usage sections and their order, e.g. to show SectionExamples last.
	// When empty, the MainUsageTemplate order is used.
	UsageSections []string
	// TipsHelp replaces the default `Use "<root> <command> --help" ...` tip.
	TipsHelp string
	// Footer is shown at the end of the root command help, e.g. a docs link.
	Footer string
	// SingleCommandList lists all root subcommands under the first command group heading,
	// instead of adding the commands that are not in any group under "Other Commands:".
	SingleCommandList bool
	// CommandShort overrides, by command name, the description shown in the root command list,
	// e.g. for the completion command that cobra adds on execute.
	CommandShort map[string]string
	// KeepLineBreaks shows command descriptions with their line breaks, instead of joining them into one line.
	KeepLineBreaks bool
}

func ActsAsRootCommand(cmd *cobra.Command, filters []string, cg CommandGroups, fg FlagGroups, envPrefix string) FlagExposer {
	return ActsAsRootCommandWithOptions(cmd, filters, cg, fg, envPrefix, HelpOptions{})
}

// ActsAsRootCommandWithOptions works like ActsAsRootCommand and applies opts to the help output.
// Passing a zero HelpOptions gives the same output as ActsAsRootCommand.
func ActsAsRootCommandWithOptions(cmd *cobra.Command, filters []string, cg CommandGroups, fg FlagGroups, envPrefix string, opts HelpOptions) FlagExposer {
	if cmd == nil {
		panic("nil root command")
	}

	usageTemplate := MainUsageTemplate()
	if len(opts.UsageSections) > 0 {
		usageTemplate = UsageTemplate(opts.UsageSections...)
	}

	helpTemplate := MainHelpTemplate()
	if opts.KeepLineBreaks {
		helpTemplate = KeepLineBreaksHelpTemplate()
	}

	templater := &templater{
		RootCmd:       cmd,
		UsageTemplate: usageTemplate,
		HelpTemplate:  helpTemplate,
		CommandGroups: cg,
		FlagGroups:    fg,
		Filtered:      filters,
		EnvPrefix:     envPrefix,
		Options:       opts,
	}
	cmd.SetFlagErrorFunc(templater.FlagErrorFunc())
	cmd.SilenceUsage = true
	cmd.SetUsageFunc(templater.UsageFunc())
	cmd.SetHelpFunc(templater.HelpFunc())
	return templater
}

func UseOptionsTemplates(cmd *cobra.Command) {
	templater := &templater{
		UsageTemplate: OptionsUsageTemplate(),
		HelpTemplate:  OptionsHelpTemplate(),
	}
	cmd.SetUsageFunc(templater.UsageFunc())
	cmd.SetHelpFunc(templater.HelpFunc())
}

type templater struct {
	UsageTemplate string
	HelpTemplate  string
	RootCmd       *cobra.Command
	CommandGroups
	FlagGroups
	Filtered  []string
	EnvPrefix string
	Options   HelpOptions
}

func (templater *templater) FlagErrorFunc(exposedFlags ...string) func(*cobra.Command, error) error {
	return func(c *cobra.Command, err error) error {
		c.SilenceUsage = true
		switch c.CalledAs() {
		case "options":
			return fmt.Errorf("%s\nRun '%s' without flags.", err, c.CommandPath())
		default:
			return fmt.Errorf("%s\nSee '%s --help' for usage.", err, c.CommandPath())
		}
	}
}

func (templater *templater) ExposeFlags(cmd *cobra.Command, flags ...string) FlagExposer {
	cmd.SetUsageFunc(templater.UsageFunc(flags...))
	return templater
}

func (templater *templater) HelpFunc() func(*cobra.Command, []string) {
	return func(c *cobra.Command, s []string) {
		t := template.New("help")
		t.Funcs(templater.templateFuncs())
		template.Must(t.Parse(templater.HelpTemplate))
		out := term.NewWordWrapWriter(c.OutOrStdout(), DefaultWrapLimit)
		err := t.Execute(out, c)
		if err != nil {
			c.Println(err)
		}
	}
}

func (templater *templater) UsageFunc(exposedFlags ...string) func(*cobra.Command) error {
	return func(c *cobra.Command) error {
		t := template.New("usage")
		t.Funcs(templater.templateFuncs(exposedFlags...))
		template.Must(t.Parse(templater.UsageTemplate))
		out := term.NewWordWrapWriter(c.OutOrStdout(), DefaultWrapLimit)
		return t.Execute(out, c)
	}
}

func (templater *templater) templateFuncs(exposedFlags ...string) template.FuncMap {
	return template.FuncMap{
		"trim": func(s string) string {
			s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
			s = strings.TrimSpace(s)
			return s
		},
		"trimRight":           func(s string) string { return strings.TrimRightFunc(s, unicode.IsSpace) },
		"trimLeft":            func(s string) string { return strings.TrimLeftFunc(s, unicode.IsSpace) },
		"gt":                  cobra.Gt,
		"eq":                  cobra.Eq,
		"rpad":                rpad,
		"appendIfNotPresent":  appendIfNotPresent,
		"flagsNotIntersected": flagsNotIntersected,
		"visibleFlags":        visibleFlags,
		"flagsUsages":         templater.flagsUsages,
		"cmdGroups":           templater.cmdGroups,
		"cmdGroupsString":     templater.cmdGroupsString,
		"rootCmd":             templater.rootCmdName,
		"isRootCmd":           templater.isRootCmd,
		"optionsCmdFor":       templater.optionsCmdFor,
		"usageLine":           templater.usageLine,
		"showRootUsage": func(c *cobra.Command) bool {
			return templater.Options.RootUsage && templater.isRootCmd(c)
		},
		"persistentFlagsHeading": func() string { return templater.Options.PersistentFlagsHeading },
		"flagsUsagesNoGroups":    templater.flagsUsagesNoGroups,
		"globalFlags":            globalFlags,
		"tipsHelp":               func() string { return templater.Options.TipsHelp },
		"footer":                 func() string { return templater.Options.Footer },
		"exposed": func(c *cobra.Command) *flag.FlagSet {
			exposed := flag.NewFlagSet("exposed", flag.ContinueOnError)
			if len(exposedFlags) > 0 {
				for _, name := range exposedFlags {
					if flag := c.Flags().Lookup(name); flag != nil {
						exposed.AddFlag(flag)
					}
				}
			}
			return exposed
		},
	}
}

func (templater *templater) cmdGroups(c *cobra.Command, all []*cobra.Command) []CommandGroup {
	if len(templater.CommandGroups) > 0 && c == templater.RootCmd {
		all = filter(all, templater.Filtered...)
		groups := AddAdditionalCommands(templater.CommandGroups, "Other Commands:", all)
		if templater.Options.SingleCommandList && len(groups) > 1 {
			var cmds []*cobra.Command
			for _, g := range groups {
				cmds = append(cmds, g.Commands...)
			}
			groups = CommandGroups{{Message: groups[0].Message, Commands: cmds}}
		}
		return groups
	}
	all = filter(all, "options")
	return []CommandGroup{
		{
			Message:  "Available Commands:",
			Commands: all,
		},
	}
}

func (t *templater) cmdGroupsString(c *cobra.Command) string {
	groups := []string{}
	for _, cmdGroup := range t.cmdGroups(c, c.Commands()) {
		cmds := []string{cmdGroup.Message}
		for _, cmd := range cmdGroup.Commands {
			if cmd.IsAvailableCommand() {
				short := cmd.Short
				if s, ok := t.Options.CommandShort[cmd.Name()]; ok && t.isRootCmd(c) {
					short = s
				}
				cmds = append(cmds, "  "+rpad(cmd.Name(), cmd.NamePadding())+"   "+short)
			}
		}
		groups = append(groups, strings.Join(cmds, "\n"))
	}
	return strings.Join(groups, "\n\n")
}

func (t *templater) rootCmdName(c *cobra.Command) string {
	return t.rootCmd(c).CommandPath()
}

func (t *templater) isRootCmd(c *cobra.Command) bool {
	return t.rootCmd(c) == c
}

func (t *templater) parents(c *cobra.Command) []*cobra.Command {
	parents := []*cobra.Command{c}
	for current := c; !t.isRootCmd(current) && current.HasParent(); {
		current = current.Parent()
		parents = append(parents, current)
	}
	return parents
}

func (t *templater) rootCmd(c *cobra.Command) *cobra.Command {
	if c != nil && !c.HasParent() {
		return c
	}
	if t.RootCmd == nil {
		panic("nil root cmd")
	}
	return t.RootCmd
}

func (t *templater) optionsCmdFor(c *cobra.Command) string {
	if !c.Runnable() {
		return ""
	}
	rootCmdStructure := t.parents(c)
	for i := len(rootCmdStructure) - 1; i >= 0; i-- {
		cmd := rootCmdStructure[i]
		if _, _, err := cmd.Find([]string{"options"}); err == nil {
			return cmd.CommandPath() + " options"
		}
	}
	return ""
}

func (t *templater) usageLine(c *cobra.Command) string {
	usage := c.UseLine()
	suffix := "[flags]"
	if c.HasFlags() && !strings.Contains(usage, suffix) {
		usage += " " + suffix
	}
	return usage
}

// flagsUsages will print out the kubectl help flags
func (t *templater) flagsUsages(f *flag.FlagSet) (string, error) {
	flagBuf := new(bytes.Buffer)
	printer := NewHelpFlagPrinter(flagBuf, t.EnvPrefix, DefaultWrapLimit)

	printFs := func() {
		f.VisitAll(func(flag *flag.Flag) {
			if flag.Hidden {
				return
			}
			printer.PrintHelpFlag(flag)
		})
	}

	g := t.FlagGroups

	if len(g) == 0 {
		fmt.Fprintf(flagBuf, "Options:\n")
		printFs()
		return flagBuf.String(), nil
	}

	for i, fs := range g.splitFlagSet(f) {
		if !fs.HasAvailableFlags() {
			continue
		}

		if len(g[i].Name) > 0 {
			fmt.Fprintf(flagBuf, "%s:\n", g[i].Name)
		}
		fs.VisitAll(func(flag *flag.Flag) {
			if flag.Hidden {
				return
			}
			printer.PrintHelpFlag(flag)
		})
	}

	return flagBuf.String(), nil
}

// flagsUsagesNoGroups prints the flags without flag group names.
func (t *templater) flagsUsagesNoGroups(f *flag.FlagSet) string {
	flagBuf := new(bytes.Buffer)
	printer := NewHelpFlagPrinter(flagBuf, t.EnvPrefix, DefaultWrapLimit)
	f.VisitAll(func(flag *flag.Flag) {
		if flag.Hidden {
			return
		}
		printer.PrintHelpFlag(flag)
	})
	return flagBuf.String()
}

// globalFlags builds a new list from the persistent flags plus the local help flag
func globalFlags(c *cobra.Command) *flag.FlagSet {
	f := flag.NewFlagSet("global", flag.ContinueOnError)
	f.AddFlagSet(c.PersistentFlags())
	if h := c.Flags().Lookup("help"); h != nil {
		f.AddFlag(h)
	}
	return f
}

// getFlagFormat will output the flag format
func getFlagFormat(f *flag.Flag) string {
	format := "--%s%s%s%s\n%s%s"
	if len(f.Shorthand) > 0 {
		format = "    -%s, " + format
	} else {
		format = "    %s" + format
	}

	return format
}

func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func appendIfNotPresent(s, stringToAppend string) string {
	if strings.Contains(s, stringToAppend) {
		return s
	}
	return s + " " + stringToAppend
}

func flagsNotIntersected(l *flag.FlagSet, r *flag.FlagSet) *flag.FlagSet {
	f := flag.NewFlagSet("notIntersected", flag.ContinueOnError)
	l.VisitAll(func(flag *flag.Flag) {
		if r.Lookup(flag.Name) == nil {
			f.AddFlag(flag)
		}
	})
	return f
}

func visibleFlags(l *flag.FlagSet) *flag.FlagSet {
	hidden := "help"
	f := flag.NewFlagSet("visible", flag.ContinueOnError)
	l.VisitAll(func(flag *flag.Flag) {
		if flag.Name != hidden {
			f.AddFlag(flag)
		}
	})
	return f
}

func filter(cmds []*cobra.Command, names ...string) []*cobra.Command {
	out := []*cobra.Command{}
	for _, c := range cmds {
		if c.Hidden {
			continue
		}
		skip := false
		for _, name := range names {
			if name == c.Name() {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		out = append(out, c)
	}
	return out
}
