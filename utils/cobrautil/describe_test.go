// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
)

func TestDescribeFlagsAsPlain(t *testing.T) {
	testDescribeFlags(t, Plain, []string{
		`a=true
b=false
c=true
d=false`,
		`key=false`,
		`key=val`,
		`key=false`,
		`key=false`,
		``,
		`list=item1,item2`,
	})
}

func TestDescribeFlagsAsJSON(t *testing.T) {
	testDescribeFlags(t, JSON, []string{
		`{"a":true,"b":false,"c":true,"d":false}`,
		`{"key":false}`,
		`{"key":"val"}`,
		`{"key":false}`,
		`{"key":false}`,
		`{}`,
		`{"list":["item1","item2"]}`,
	})
}

func TestDescribeFlagsAsYAML(t *testing.T) {
	testDescribeFlags(t, YAML, []string{
		`a: true
b: false
c: true
d: false`,
		`key: false`,
		`key: val`,
		`key: false`,
		`key: false`,
		`{}`,
		`list:
  - item1
  - item2`,
	})
}

func testDescribeFlags(t *testing.T, f DescribeFormat, expected []string) { //nolint:thelper // not a helper
	tests := []struct {
		name     string
		flags    func() *pflag.FlagSet
		decorate func(*FlagsDescriber)
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
		},
		{
			name: "bool is correctly formatted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				return fs
			},
		},
		{
			name: "string is correctly formatted",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.String("key", "val", "")
				return fs
			},
		},
		{
			name: "help is not shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				fs.Bool("help", true, "")
				return fs
			},
		},
		{
			name: "hidden is shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				_ = fs.MarkHidden("key")
				return fs
			},
			decorate: func(d *FlagsDescriber) {
				d.ShowHidden = true
			},
		},
		{
			name: "hidden is not shown",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.Bool("key", false, "")
				_ = fs.MarkHidden("key")
				return fs
			},
		},
		{
			name: "list of values",
			flags: func() *pflag.FlagSet {
				fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)
				fs.StringSlice("list", []string{"item1", "item2"}, "")
				return fs
			},
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			d := FlagsDescriber{
				Format: f,
			}
			if tc.decorate != nil {
				tc.decorate(&d)
			}

			result, err := d.DescribeFlags(tc.flags())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(expected[i], strings.TrimSpace(result)); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
