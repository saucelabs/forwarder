// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

func isNullOrUndefined(v goja.Value) bool {
	return v == nil || goja.IsUndefined(v) || goja.IsNull(v)
}

func asString(v goja.Value) (string, bool) {
	if v == nil {
		return "", false
	}
	s, ok := v.Export().(string)
	return s, ok
}

func asSlice[T any](s, delim string, parse func(v string) (T, error)) ([]T, error) {
	l := strings.Split(s, delim)
	res := make([]T, 0, len(l))
	for i, v := range l {
		// Skip empty values.
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		// Parse the value.
		r, err := parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q at pos %d: %w", v, i, err)
		}

		res = append(res, r)
	}

	return res, nil
}

func semicolonDelimitedString[T fmt.Stringer](values []T) string {
	if len(values) == 0 {
		return ""
	}

	s := make([]string, len(values))
	for i, ip := range values {
		s[i] = ip.String()
	}
	return strings.Join(s, ";")
}
