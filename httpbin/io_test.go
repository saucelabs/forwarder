// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package httpbin

import (
	"bytes"
	"io"
	"testing"
)

func TestPatternReader(t *testing.T) {
	r := &patternReader{
		Pattern: []byte("0123456789"),
		N:       int64(75),
	}
	var buf bytes.Buffer
	n, err := io.Copy(&buf, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 75 {
		t.Fatalf("n=%d, want 75", n)
	}
	t.Log(buf.String())
}
