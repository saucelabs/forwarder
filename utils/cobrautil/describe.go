// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
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

	"github.com/saucelabs/forwarder/utils/maps"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type DescribeFormat int

const (
	Plain DescribeFormat = iota
	JSON
	YAML
	OneLine
)

type FlagsDescriber struct {
	Format          DescribeFormat
	Unredacted      bool
	ShowChangedOnly bool
	ShowHidden      bool
}

func (d FlagsDescriber) DescribeFlags(fs *pflag.FlagSet) ([]byte, error) {
	args := d.DescribeFlagsToMap(fs)

	switch d.Format {
	case Plain:
		keys := maps.Keys(args)
		sort.Strings(keys)
		var buf bytes.Buffer
		for _, name := range keys {
			buf.WriteString(fmt.Sprintf("%s=%s\n", name, args[name]))
		}
		return buf.Bytes(), nil
	case JSON:
		return json.Marshal(args)
	case YAML:
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(args); err != nil {
			return nil, err
		}
		if err := enc.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case OneLine:
		keys := maps.Keys(args)
		sort.Strings(keys)
		var buf bytes.Buffer
		for i, name := range keys {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%s=%s", name, args[name]))
		}
		return buf.Bytes(), nil
	default:
		return nil, errors.New("unknown format")
	}
}

func (d FlagsDescriber) DescribeFlagsToMap(fs *pflag.FlagSet) map[string]any {
	args := make(map[string]any, fs.NFlag())

	fs.VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		if !f.Changed && d.ShowChangedOnly {
			return
		}
		if f.Hidden && !d.ShowHidden {
			return
		}

		val := f.Value
		if d.Unredacted {
			if v, ok := f.Value.(redactedValue); ok {
				val = v.Unredacted()
			}
		}

		if val.Type() == "bool" {
			args[f.Name] = val
		} else {
			if sv, ok := val.(sliceValue); ok {
				if d.Format == Plain {
					args[f.Name] = strings.Join(sv.GetSlice(), ",")
				} else {
					args[f.Name] = sv.GetSlice()
				}
			} else {
				args[f.Name] = val.String()
			}
		}
	})

	return args
}

type sliceValue interface {
	GetSlice() []string
}

type redactedValue interface {
	Unredacted() pflag.Value
}
