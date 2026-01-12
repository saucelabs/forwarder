// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"fmt"
)

type NamedParam[T fmt.Stringer] struct {
	Name  string
	Param *T
}

func (p NamedParam[T]) String() string {
	if p.Name == "" {
		return (*p.Param).String()
	}

	return p.Name + ":" + (*p.Param).String()
}
