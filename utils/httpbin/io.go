// Copyright 2022-2026 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpbin

import (
	"io"
)

type patternReader struct {
	Pattern []byte
	N       int64
}

func (l *patternReader) Read(p []byte) (int, error) {
	if l.N <= 0 {
		return 0, io.EOF
	}

	if int64(len(p)) > l.N {
		p = p[:l.N]
	}
	var n int
	for {
		n0 := copy(p, l.Pattern)
		n += n0
		if n0 < len(l.Pattern) {
			break
		}
		p = p[n0:]
	}

	l.N -= int64(n)
	return n, nil
}
