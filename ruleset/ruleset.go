// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ruleset

import (
	"errors"
	"regexp"
	"strings"
)

type Regexp struct {
	include *regexp.Regexp
	exclude *regexp.Regexp
}

var ErrNoIncludeRules = errors.New("no include rules specified")

// NewRegexp returns the Regexp with given include and exclude rules.
func NewRegexp(include, exclude []*regexp.Regexp) (*Regexp, error) {
	if len(include) == 0 {
		return nil, ErrNoIncludeRules
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

	return &Regexp{
		include: build(include),
		exclude: build(exclude),
	}, nil
}

// Match returns true if the given string matches at least one of the include rules
// and does not match the exclude rules.
func (r *Regexp) Match(s string) bool {
	if r.exclude != nil && r.exclude.MatchString(s) {
		return false
	}
	return r.include != nil && r.include.MatchString(s)
}

// ParseRegexp parses the given rules into a Regexp.
// Rules that start with "-" are treated as exclude rules.
func ParseRegexp(in []string) (*Regexp, error) {
	var include, exclude []*regexp.Regexp
	for _, rule := range in {
		rule, ex := strings.CutPrefix(rule, "-")
		reg, err := regexp.Compile(rule)
		if err != nil {
			return nil, err
		}

		if ex {
			exclude = append(exclude, reg)
		} else {
			include = append(include, reg)
		}
	}

	return NewRegexp(include, exclude)
}
