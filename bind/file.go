// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"os"

	"github.com/spf13/pflag"
)

// osFileFlag allows to print the file name instead of the file descriptor.
type osFileFlag struct {
	pflag.Value
	f **os.File
}

func (f *osFileFlag) String() string {
	if *f.f == nil {
		return ""
	}
	return (*f.f).Name()
}

func newOSFileFlag(v pflag.Value, f **os.File) pflag.Value {
	if f == nil {
		panic("nil pointer")
	}
	return &osFileFlag{v, f}
}
