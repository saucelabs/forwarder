// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"os"

	"github.com/mmatczuk/anyflag"
	"github.com/spf13/pflag"
)

// FileFlag wraps anyflag.NewValue[*os.File] to fix the String() method.
// When the file is nil, it returns an empty string.
type FileFlag struct {
	anyflag.Value[*os.File]
	f **os.File
}

func (f *FileFlag) String() string {
	if *f.f == nil {
		return ""
	}
	return (*f.f).Name()
}

func NewFileFlag(f **os.File, p func(val string) (*os.File, error)) pflag.Value {
	if f == nil {
		panic("nil pointer")
	}

	v := anyflag.NewValue[*os.File](*f, f, p)
	return &FileFlag{*v, f}
}
