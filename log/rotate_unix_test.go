// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build unix

package log

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestRotatableFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-rotate-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	f, err := os.OpenFile(path.Join(dir, "0.log"), DefaultFileFlags, DefaultFileMode)
	if err != nil {
		t.Fatal(err)
	}
	r := NewRotatableFile(f)
	defer r.Close()

	for i := 1; i < 10; i++ {
		_, err := r.Write([]byte("hello\n"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(f.Name(), path.Join(dir, fmt.Sprintf("%d.log", i))); err != nil {
			t.Fatal(err)
		}
		if err := r.Reopen(); err != nil {
			t.Fatal(err)
		}
	}

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 1; i < 10; i++ {
		b, err := os.ReadFile(path.Join(dir, fmt.Sprintf("%d.log", i)))
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != "hello\n" {
			t.Fatalf("unexpected content: %q at file %d", b, i)
		}
	}
}
