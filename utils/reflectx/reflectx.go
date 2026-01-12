// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package reflectx

import (
	"reflect"
)

// LookupImpl performs a depth-first search for a value that implements T.
func LookupImpl[T any](v reflect.Value) (T, bool) {
	var nop T

	if v.CanInterface() {
		if impl, ok := v.Interface().(T); ok {
			return impl, true
		}

		// This works around issues with embedded interfaces.
		v = reflect.ValueOf(v.Interface())
	}

	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return nop, false
	}

	for i := range v.NumField() {
		f := v.Field(i)

		if f.CanInterface() {
			if impl, ok := f.Interface().(T); ok {
				return impl, true
			}
		}
	}

	for i := range v.NumField() {
		if impl, ok := LookupImpl[T](v.Field(i)); ok {
			return impl, true
		}
	}

	return nop, false
}
