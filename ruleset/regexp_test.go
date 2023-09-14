// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ruleset

import (
	"errors"
	"regexp"
	"testing"
)

func TestRuleSet(t *testing.T) {
	tests := []struct {
		name          string
		include       []*regexp.Regexp
		exclude       []*regexp.Regexp
		match         []string
		dontMatch     []string
		expectedError error
	}{
		{
			name:    "include all",
			include: []*regexp.Regexp{regexp.MustCompile(".*")},
			match:   []string{"foo", "bar"},
		},
		{
			name:      "exclude all",
			include:   []*regexp.Regexp{regexp.MustCompile("")},
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
			include:   []*regexp.Regexp{regexp.MustCompile(".*")},
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
			name:    "many excludes",
			include: []*regexp.Regexp{regexp.MustCompile(".*")},
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
		{
			name:          "no includes",
			expectedError: ErrNoIncludeRules,
		},
		{
			name:          "no includes with excludes",
			exclude:       []*regexp.Regexp{regexp.MustCompile(".*")},
			expectedError: ErrNoIncludeRules,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			rs, err := NewRegexp(tc.include, tc.exclude)
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("expected error %v, got %v", tc.expectedError, err)
			}
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
