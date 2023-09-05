// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

type DescribeFormat int

const (
	Plain DescribeFormat = iota
	JSON
)

func DescribeFlags(fs *pflag.FlagSet, showHidden bool, format DescribeFormat) (string, error) {
	args := make(map[string]any, fs.NFlag())
	keys := make([]string, 0, fs.NFlag())

	fs.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}

		if flag.Hidden && !showHidden {
			return
		}

		if flag.Value.Type() == "bool" {
			args[flag.Name] = flag.Value
		} else {
			args[flag.Name] = strings.Trim(flag.Value.String(), "[]")
		}

		keys = append(keys, flag.Name)
	})

	sort.Strings(keys)

	switch format {
	case Plain:
		var b strings.Builder
		for _, name := range keys {
			b.WriteString(fmt.Sprintf("%s=%s\n", name, args[name]))
		}
		return b.String(), nil
	case JSON:
		encoded, err := json.Marshal(args)
		return string(encoded), err
	default:
		return "", fmt.Errorf("unknown format requested")
	}
}
