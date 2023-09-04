// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ruleset

import (
	"regexp"
	"strings"
)

type RuleSet struct {
	include *regexp.Regexp
	exclude *regexp.Regexp
}

// NewRuleSet returns the RuleSet with given include and exclude rules.
// When only the exclude rules are specified, the include rule is set to match everything.
func NewRuleSet(include, exclude []*regexp.Regexp) *RuleSet {
	if len(include) == 0 && len(exclude) != 0 {
		include = []*regexp.Regexp{regexp.MustCompile(".*")}
	}

	build := func(rules []*regexp.Regexp) *regexp.Regexp {
		var regex strings.Builder
		for i := range rules {
			if i > 0 {
				regex.WriteString("|")
			}
			regex.WriteString(rules[i].String())
		}
		if s := regex.String(); s != "" {
			return regexp.MustCompile(s)
		}
		return nil
	}

	return &RuleSet{
		include: build(include),
		exclude: build(exclude),
	}
}

// Match returns true if the given string matches at least one of the include rules
// and does not match the exclude rules.
func (r *RuleSet) Match(s string) bool {
	if r.exclude != nil && r.exclude.MatchString(s) {
		return false
	}
	return r.include != nil && r.include.MatchString(s)
}
