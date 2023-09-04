// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ruleset

import (
	"regexp"
	"testing"
)

func TestRuleSet(t *testing.T) {
	tests := []struct {
		name      string
		include   []*regexp.Regexp
		exclude   []*regexp.Regexp
		match     []string
		dontMatch []string
	}{
		{
			name:      "empty",
			dontMatch: []string{"foo", "bar"},
		},
		{
			name:    "include all",
			include: []*regexp.Regexp{regexp.MustCompile(".*")},
			match:   []string{"foo", "bar"},
		},
		{
			name:      "exclude all",
			exclude:   []*regexp.Regexp{regexp.MustCompile(".*")},
			dontMatch: []string{"foo", "bar"},
		},
		{
			name:      "include foo",
			include:   []*regexp.Regexp{regexp.MustCompile("foo")},
			match:     []string{"foo"},
			dontMatch: []string{"bar"},
		},
		{
			name:      "exclude foo",
			exclude:   []*regexp.Regexp{regexp.MustCompile("foo")},
			match:     []string{"bar"},
			dontMatch: []string{"foo"},
		},
		{
			name: "many includes",
			include: []*regexp.Regexp{
				regexp.MustCompile("foo"),
				regexp.MustCompile("bar"),
				regexp.MustCompile("baz"),
			},
			match:     []string{"foo", "bar", "baz", "foobar"},
			dontMatch: []string{"aa", "bb"},
		},
		{
			name: "many excludes",
			exclude: []*regexp.Regexp{
				regexp.MustCompile("foo"),
				regexp.MustCompile("bar"),
				regexp.MustCompile("baz"),
			},
			match:     []string{"aa", "bb"},
			dontMatch: []string{"foo", "bar", "baz", "foobar"},
		},
		{
			name: "includes and excludes",
			include: []*regexp.Regexp{
				regexp.MustCompile("foo"),
				regexp.MustCompile("bar"),
			},
			exclude: []*regexp.Regexp{
				regexp.MustCompile("fooo"),
				regexp.MustCompile("bark"),
			},
			match:     []string{"foo", "bar", "foobar"},
			dontMatch: []string{"fooo", "bark", "foobarkey"},
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			rs := NewRuleSet(tc.include, tc.exclude)
			for _, m := range tc.match {
				if !rs.Match(m) {
					t.Errorf("expected %q to match", m)
				}
			}
			for _, m := range tc.dontMatch {
				if rs.Match(m) {
					t.Errorf("expected %q to not match", m)
				}
			}
		})
	}
}
