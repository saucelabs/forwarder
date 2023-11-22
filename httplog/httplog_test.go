// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httplog

import (
	"testing"
)

func TestSplitNameMode(t *testing.T) {
	tests := []struct {
		val  string
		name string
		mode Mode
	}{
		{
			val:  "api:none",
			name: "api",
			mode: None,
		},
		{
			val:  "errors",
			name: "",
			mode: Errors,
		},
	}

	for _, tc := range tests {
		t.Run(tc.val, func(t *testing.T) {
			n, m, err := SplitNameMode(tc.val)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if n != tc.name {
				t.Errorf("expected name %q, got %q", tc.name, n)
			}
			if m != tc.mode {
				t.Errorf("expected mode %q, got %q", tc.mode, m)
			}
		})
	}
}

func TestSplitNameModeError(t *testing.T) {
	tests := []string{
		"api:invalid",
		"invalid",
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			_, _, err := SplitNameMode(tc)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			t.Log(err)
		})
	}
}
