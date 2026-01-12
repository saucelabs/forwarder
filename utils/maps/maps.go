// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package maps

import (
	"maps"
	"slices"
)

// Keys is a short alias to preserve the behavior of "golang.org/x/exp/maps.Keys",
// which is now flagged by linters as outdated. The suggested replacement is more verbose
// and, in my opinion, less readable. This alias provides a concise and readable alternative.
func Keys[Map ~map[K]V, K comparable, V any](m Map) []K {
	return slices.AppendSeq(make([]K, 0, len(m)), maps.Keys(m))
}

// Copy is there to reduce imports, when using both Keys and Copy.
func Copy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	maps.Copy(dst, src)
}
