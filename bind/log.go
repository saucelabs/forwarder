// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"strings"

	"github.com/mmatczuk/anyflag"
	"github.com/saucelabs/forwarder/httplog"
)

type httplogFlag struct {
	*anyflag.SliceValue[NamedParam[httplog.Mode]]
	update func()
}

func (f httplogFlag) Set(val string) (err error) {
	err = f.SliceValue.Set(val)

	if err == nil {
		f.update()
	}

	return
}

func (f httplogFlag) Replace(vals []string) (err error) {
	err = f.SliceValue.Replace(vals)

	if err == nil {
		f.update()
	}

	return
}

func (f httplogFlag) String() string {
	s := f.SliceValue.GetSlice()
	if len(s) == 0 {
		return httplog.DefaultMode.String()
	}

	// Check if all modes are the same.
	// If not, print all (name,mode) pairs.
	for i := 1; i < len(s); i++ {
		_, m1, _ := strings.Cut(s[i-1], ":")
		if m1 == "" {
			m1 = s[i-1]
		}
		_, m2, _ := strings.Cut(s[i], ":")
		if m2 == "" {
			m2 = s[i]
		}
		if m1 != m2 {
			return f.SliceValue.String()
		}
	}

	// All modes are the same - print only the mode.
	_, m, _ := strings.Cut(s[0], ":")
	if m == "" {
		m = s[0]
	}

	return m
}

func httplogUpdate(dst, src []NamedParam[httplog.Mode]) {
	changed := make([]bool, len(dst))

	// Update dst with src values.
	for i := range dst {
		for j := range src {
			if dst[i].Name == src[j].Name {
				*dst[i].Param = *src[j].Param
				changed[i] = true
				break
			}
		}
	}

	// Find default mode.
	var defaultMode httplog.Mode
	for i := range src {
		j := len(src) - i - 1
		if src[j].Name == "" {
			defaultMode = *src[j].Param
			break
		}
	}

	// If default mode is set, update dst with it.
	if defaultMode != "" {
		for i := range dst {
			if !changed[i] {
				*dst[i].Param = defaultMode
			}
		}
	}
}

func httplogExtractNames(cfg []NamedParam[httplog.Mode]) []string {
	names := make([]string, 0, len(cfg))
	for _, c := range cfg {
		if c.Name != "" {
			names = append(names, c.Name)
		}
	}
	return names
}
