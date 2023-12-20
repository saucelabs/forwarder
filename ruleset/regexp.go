// Copyright 2023 Sauce Labs Inc., all rights reserved.
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

type RegexpMatcher struct {
	include *regexp.Regexp
	exclude *regexp.Regexp
	inverse bool
}

var ErrNoIncludeRules = errors.New("no include rules specified")

// NewRegexpMatcher returns the RegexpMatcher with given include and exclude rules.
func NewRegexpMatcher(include, exclude []*regexp.Regexp) (*RegexpMatcher, error) {
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

	return &RegexpMatcher{
		include: build(include),
		exclude: build(exclude),
	}, nil
}

// Inverse returns a new RegexpMatcher that inverts the match result.
func (r *RegexpMatcher) Inverse() *RegexpMatcher {
	return &RegexpMatcher{
		include: r.include,
		exclude: r.exclude,
		inverse: !r.inverse,
	}
}

// Match returns true if the given string matches at least one of the include rules
// and does not match the exclude rules.
func (r *RegexpMatcher) Match(s string) bool {
	m := r.match(s)
	if r.inverse {
		m = !m
	}
	return m
}

func (r *RegexpMatcher) match(s string) bool {
	if r.exclude != nil && r.exclude.MatchString(s) {
		return false
	}
	return r.include != nil && r.include.MatchString(s)
}

type RegexpListItem struct {
	*regexp.Regexp
	Exclude bool
}

func ParseRegexpListItem(val string) (RegexpListItem, error) {
	val, exclude := strings.CutPrefix(val, "-")
	r, err := regexp.Compile(val)
	if err != nil {
		return RegexpListItem{}, err
	}
	return RegexpListItem{r, exclude}, nil
}

func (r RegexpListItem) String() string {
	if r.Exclude {
		return "-" + r.Regexp.String()
	}
	return r.Regexp.String()
}

func NewRegexpMatcherFromList(l []RegexpListItem) (*RegexpMatcher, error) {
	var include, exclude []*regexp.Regexp
	for i := range l {
		if l[i].Exclude {
			exclude = append(exclude, l[i].Regexp)
		} else {
			include = append(include, l[i].Regexp)
		}
	}
	return NewRegexpMatcher(include, exclude)
}
