// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type DescribeFormat int

const (
	Plain DescribeFormat = iota
	JSON
	YAML
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

	fs.VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		if f.Hidden && !d.ShowHidden {
			return
		}

		if f.Value.Type() == "bool" {
			args[f.Name] = f.Value
		} else {
			if sv, ok := f.Value.(sliceValue); ok {
				if d.Format == Plain {
					args[f.Name] = strings.Join(sv.GetSlice(), ",")
				} else {
					args[f.Name] = sv.GetSlice()
				}
			} else {
				args[f.Name] = f.Value.String()
			}
		}
	})

	switch d.Format {
	case Plain:
		keys := maps.Keys(args)
		sort.Strings(keys)
		var sb strings.Builder
		for _, name := range keys {
			sb.WriteString(fmt.Sprintf("%s=%s\n", name, args[name]))
		}
		return sb.String(), nil
	case JSON:
		b, err := json.Marshal(args)
		return string(b), err
	case YAML:
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(args); err != nil {
			return "", err
		}
		if err := enc.Close(); err != nil {
			return "", err
		}
		return buf.String(), nil
	default:
		return "", errors.New("unknown format")
	}
}

type sliceValue interface {
	GetSlice() []string
}
