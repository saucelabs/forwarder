// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bind

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestDescribeFlagsAsPlain(t *testing.T) {
	tests := map[string]struct {
		input      map[string]interface{}
		expected   string
		isErr      bool
		isHidden   bool
		showHidden bool
	}{
		"keys are sorted": {
			input:    map[string]interface{}{"foo": false, "bar": true},
			expected: "bar=true\nfoo=false\n",
			isErr:    false,
		},
		"bool is correctly formatted": {
			input:    map[string]interface{}{"key": false},
			expected: "key=false\n",
			isErr:    false,
		},
		"string is correctly formatted": {
			input:    map[string]interface{}{"key": "val"},
			expected: "key=val\n",
			isErr:    false,
		},
		"help is not shown": {
			input:    map[string]interface{}{"key": false, "help": true},
			expected: "key=false\n",
			isErr:    false,
		},
		"hidden is shown": {
			input:      map[string]interface{}{"key": false},
			expected:   "key=false\n",
			isErr:      false,
			isHidden:   true,
			showHidden: true,
		},
		"hidden is not shown": {
			input:      map[string]interface{}{"key": false},
			expected:   ``,
			isErr:      false,
			isHidden:   true,
			showHidden: false,
		},
	}

	for name, tc := range tests {
		fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)

		for k, v := range tc.input {
			switch val := v.(type) {
			case bool:
				fs.Bool(k, val, "")
			case string:
				fs.String(k, val, "")
			}

			if tc.isHidden {
				err := fs.MarkHidden(k)
				if err != nil {
					t.Errorf("%s: test setup failed: %s", name, err)
				}
			}
		}
		result, err := DescribeFlags(fs, tc.showHidden, Plain)

		if (err != nil) != tc.isErr {
			t.Errorf("%s: expected error: %v, got %s", name, tc.isErr, err)
		}

		if result != tc.expected {
			t.Errorf("%s: expected %s, got %s", name, tc.expected, result)
		}
	}
}

func TestDescribeFlagsAsJSON(t *testing.T) {
	tests := map[string]struct {
		input      map[string]interface{}
		expected   string
		isErr      bool
		isHidden   bool
		showHidden bool
	}{
		"bool is not quoted": {
			input:    map[string]interface{}{"key": false},
			expected: `{"key":false}`,
			isErr:    false,
		},
		"help is not shown": {
			input:    map[string]interface{}{"key": false, "help": true},
			expected: `{"key":false}`,
			isErr:    false,
		},
		"hidden is shown": {
			input:      map[string]interface{}{"key": false},
			expected:   `{"key":false}`,
			isErr:      false,
			isHidden:   true,
			showHidden: true,
		},
		"hidden is not shown": {
			input:      map[string]interface{}{"key": false},
			expected:   `{}`,
			isErr:      false,
			isHidden:   true,
			showHidden: false,
		},
		"string is quoted": {
			input:    map[string]interface{}{"key": "val"},
			expected: `{"key":"val"}`,
			isErr:    false,
		},
	}

	for name, tc := range tests {
		fs := pflag.NewFlagSet("flags", pflag.ContinueOnError)

		for k, v := range tc.input {
			switch val := v.(type) {
			case bool:
				fs.Bool(k, val, "")
			case string:
				fs.String(k, val, "")
			}

			if tc.isHidden {
				err := fs.MarkHidden(k)
				if err != nil {
					t.Errorf("%s: test setup failed: %s", name, err)
				}
			}
		}
		result, err := DescribeFlags(fs, tc.showHidden, JSON)

		if (err != nil) != tc.isErr {
			t.Errorf("%s: expected error: %v, got %s", name, tc.isErr, err)
		}

		if result != tc.expected {
			t.Errorf("%s: expected %s, got %s", name, tc.expected, result)
		}
	}
}
