// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
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

	// Find default mode if set.
	defaultMode := httplog.Errors
	for i := range src {
		j := len(src) - i - 1
		if src[j].Name == "" {
			defaultMode = *src[j].Param
			break
		}
	}

	// Set default mode for unset values.
	for i := range dst {
		if !changed[i] {
			*dst[i].Param = defaultMode
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
