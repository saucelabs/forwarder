// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

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
