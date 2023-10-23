// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/exp/maps"
)

type DescribeFormat int

const (
	Plain DescribeFormat = iota
	JSON
)

func DescribeFlags(fs *pflag.FlagSet, format DescribeFormat) (string, error) {
	return FlagsDescriber{
		Format: format,
	}.DescribeFlags(fs)
}

type FlagsDescriber struct {
	Format     DescribeFormat
	ShowHidden bool
}

func (d FlagsDescriber) DescribeFlags(fs *pflag.FlagSet) (string, error) {
	args := make(map[string]any, fs.NFlag())

	fs.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}
		if flag.Hidden && !d.ShowHidden {
			return
		}

		if flag.Value.Type() == "bool" {
			args[flag.Name] = flag.Value
		} else {
			args[flag.Name] = strings.Trim(flag.Value.String(), "[]")
		}
	})

	switch d.Format {
	case Plain:
		keys := maps.Keys(args)
		sort.Strings(keys)
		var b strings.Builder
		for _, name := range keys {
			b.WriteString(fmt.Sprintf("%s=%s\n", name, args[name]))
		}
		return b.String(), nil
	case JSON:
		encoded, err := json.Marshal(args)
		return string(encoded), err
	default:
		return "", errors.New("unknown format")
	}
}
