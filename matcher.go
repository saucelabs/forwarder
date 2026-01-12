// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

type Matcher interface {
	Match(string) bool
}

type MatchFunc func(string) bool

func (m MatchFunc) Match(s string) bool {
	return m(s)
}
