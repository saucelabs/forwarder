// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package templates

import (
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

type FlagGroup struct {
	Name   string
	Prefix []string
}

type FlagGroups []FlagGroup

type prefixFlagSet struct {
	value string
	fs    *pflag.FlagSet
}

// splitFlagSet splits a flag set into multiple flag sets based on the prefix of the flag names.
// If multiple groups match a flag, the flag is added to the first matching group.
// The returned flag sets are ordered by the order of the groups.
func (g FlagGroups) splitFlagSet(f *pflag.FlagSet) []*pflag.FlagSet {
	var result []*pflag.FlagSet
	for _, p := range g {
		result = append(result, pflag.NewFlagSet(p.Name, pflag.ExitOnError))
	}

	// Sort the groups by the length of the prefix, so that longer prefixes are matched first.
	prefix := make([]prefixFlagSet, 0, len(g))
	for i := range g {
		for _, p := range g[i].Prefix {
			prefix = append(prefix, prefixFlagSet{p, result[i]})
		}
	}
	slices.SortFunc(prefix, func(a, b prefixFlagSet) int {
		return len(b.value) - len(a.value)
	})

	f.VisitAll(func(f *pflag.Flag) {
		for i := range prefix {
			if strings.HasPrefix(f.Name, prefix[i].value) {
				prefix[i].fs.AddFlag(f)
				break
			}
		}
	})

	return result
}

func SplitFlagSet(g FlagGroups, f *pflag.FlagSet) []*pflag.FlagSet {
	return g.splitFlagSet(f)
}
