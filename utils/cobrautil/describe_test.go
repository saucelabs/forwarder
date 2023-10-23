// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestDescribeFlagsAsPlain(t *testing.T) {
	tests := []struct {
		name       string
		flags      func() *pflag.FlagSet
		showHidden bool
		expected   string
	}{
		{
			name: "keys are sorted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("c", true, "")
				fs.Bool("d", false, "")
				fs.Bool("a", true, "")
				fs.Bool("b", false, "")
				return fs
			},
			showHidden: false,
			expected:   "a=true\nb=false\nc=true\nd=false\n",
		},
		{
			name: "bool is correctly formatted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				return fs
			},
			showHidden: false,
			expected:   "key=false\n",
		},
		{
			name: "string is correctly formatted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.String("key", "val", "")
				return fs
			},
			showHidden: false,
			expected:   "key=val\n",
		},
		{
			name: "help is not shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				fs.Bool("help", true, "")
				return fs
			},
			showHidden: false,
			expected:   "key=false\n",
		},
		{
			name: "hidden is shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				_ = fs.MarkHidden("key")
				return fs
			},
			showHidden: true,
			expected:   "key=false\n",
		},
		{
			name: "hidden is not shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				_ = fs.MarkHidden("key")
				return fs
			},
			showHidden: false,
			expected:   ``,
		},
		{
			name: "list of values",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.StringSlice("list", []string{"item1", "item2"}, "")
				return fs
			},
			showHidden: false,
			expected:   "list=item1,item2\n",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			d := FlagsDescriber{
				Format:     Plain,
				ShowHidden: tc.showHidden,
			}

			result, err := d.DescribeFlags(tc.flags())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestDescribeFlagsAsJSON(t *testing.T) {
	tests := []struct {
		name       string
		flags      func() *pflag.FlagSet
		showHidden bool
		expected   string
	}{
		{
			name: "bool is not quoted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				return fs
			},
			showHidden: false,
			expected:   `{"key":false}`,
		},
		{
			name: "help is not shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				fs.Bool("help", true, "")
				return fs
			},
			showHidden: false,
			expected:   `{"key":false}`,
		},
		{
			name: "hidden is shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				_ = fs.MarkHidden("key")
				return fs
			},
			showHidden: true,
			expected:   `{"key":false}`,
		},
		{
			name: "hidden is not shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				_ = fs.MarkHidden("key")
				return fs
			},
			showHidden: false,
			expected:   `{}`,
		},
		{
			name: "string is quoted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.String("key", "val", "")
				return fs
			},
			showHidden: false,
			expected:   `{"key":"val"}`,
		},
		{
			name: "keys are sorted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("c", false, "")
				fs.String("b", "val", "")
				fs.Bool("a", false, "")
				fs.String("d", "val", "")
				return fs
			},
			showHidden: false,
			expected:   `{"a":false,"b":"val","c":false,"d":"val"}`,
		},
		{
			name: "list of values",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.StringSlice("list", []string{"item1", "item2"}, "")
				return fs
			},
			showHidden: false,
			expected:   `{"list":["item1","item2"]}`,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			d := FlagsDescriber{
				Format:     JSON,
				ShowHidden: tc.showHidden,
			}

			result, err := d.DescribeFlags(tc.flags())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}
