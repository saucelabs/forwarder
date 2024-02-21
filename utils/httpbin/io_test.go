// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
