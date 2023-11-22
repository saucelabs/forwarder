// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package header

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected Header
	}{
		{
			input: "-RemoveMe",
			expected: Header{
				Name:   "RemoveMe",
				Action: Remove,
			},
		},
		{
			input: "-RemoveMeByPrefix*",
			expected: Header{
				Name:   "RemoveMeByPrefix",
				Action: RemoveByPrefix,
			},
		},
		{
			input: "EmptyMe;",
			expected: Header{
				Name:   "EmptyMe",
				Action: Empty,
			},
		},
		{
			input: "AddMe:value",
			expected: Header{
				Name:   "AddMe",
				Action: Add,
				Value:  &([]string{"value"})[0],
			},
		},
		{
			input: "AddMe: value",
			expected: Header{
				Name:   "AddMe",
				Action: Add,
				Value:  &([]string{"value"})[0],
			},
		},
		{
			input: "AddMe: value: value",
			expected: Header{
				Name:   "AddMe",
				Action: Add,
				Value:  &([]string{"value: value"})[0],
			},
		},
		{
			input: `AddMe: value
`,
			expected: Header{
				Name:   "AddMe",
				Action: Add,
				Value:  &([]string{"value"})[0],
			},
		},
	}
	for i := range tests {
		tc := &tests[i]
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseHeader(tc.input)
			if err != nil {
				t.Errorf("ParseHeader() error = %v", err)
			}
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("ParseHeader() diff = %v", diff)
			}
		})
	}
}

func TestParseHeaderError(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   error
	}{
		{
			name:  "empty",
			input: "",
		},
		{
			name:  "remove invalid name",
			input: "-(@Me)",
		},
		{
			name:  "add invalid name",
			input: "@Me: value",
		},
		{
			name: "add invalid value",
			input: `AddMe: value
value2`,
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseHeader(tc.input)
			t.Log(err)
			if err == nil {
				t.Errorf("ParseHeader() error = %v", err)
			}
		})
	}
}

func TestRemoveHeadersByPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		header   http.Header
		expected http.Header
	}{
		{
			name:   "smoke",
			prefix: http.CanonicalHeaderKey("RemoveMe"),
			header: http.Header{
				http.CanonicalHeaderKey("Remo"):             nil,
				http.CanonicalHeaderKey("RemoveMeByPrefix"): nil,
				http.CanonicalHeaderKey("RemoveMeBy"):       nil,
				http.CanonicalHeaderKey("RemoveMe"):         nil,
				http.CanonicalHeaderKey("DontRemoveMe"):     nil,
			},
			expected: http.Header{
				http.CanonicalHeaderKey("Remo"):         nil,
				http.CanonicalHeaderKey("DontRemoveMe"): nil,
			},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			h := tc.header.Clone()
			removeHeadersByPrefix(h, tc.prefix)

			if diff := cmp.Diff(h, tc.expected); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
