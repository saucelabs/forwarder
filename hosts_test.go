// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import "testing"

func TestReadHostsFile(t *testing.T) {
	m, err := ReadHostsFile()
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range m {
		t.Logf("%s=%s", k, v)
	}
	if len(m) == 0 {
		t.Fatal("no hosts found")
	}
	if m["localhost"] == nil {
		t.Fatal("localhost not found")
	}
}
