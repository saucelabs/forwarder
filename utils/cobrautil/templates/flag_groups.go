// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package templates

import (
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

type FlagGroup struct {
	Name   string
	Prefix string

	priority int
}

type FlagGroups []FlagGroup

// processedFlagGroups returns a copy of the flag groups with the priority field set to the index of the group.
// The returned groups are sorted by the length of the prefix in descending order.
func processedFlagGroups(g FlagGroups) FlagGroups {
	c := make(FlagGroups, len(g))
	copy(c, g)

	for i := range c {
		c[i].priority = i
	}

	sort.Slice(c, func(i, j int) bool {
		return len(c[i].Prefix) > len(c[j].Prefix)
	})

	return c
}

// splitFlagSet splits a flag set into multiple flag sets based on the prefix of the flag names.
// If multiple groups match a flag, the flag is added to the first matching group.
// The returned flag sets are ordered by the order of the groups.
func (g FlagGroups) splitFlagSet(f *pflag.FlagSet) []*pflag.FlagSet {
	var result []*pflag.FlagSet
	for _, p := range g {
		result = append(result, pflag.NewFlagSet(p.Name, pflag.ExitOnError))
	}

	f.VisitAll(func(f *pflag.Flag) {
		for i := range g {
			if strings.HasPrefix(f.Name, g[i].Prefix) {
				result[i].AddFlag(f)
				break
			}
		}
	})
	return result
}
