// Copyright 2021 The randomness Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package randomness

import "testing"

func Test_randomPortGenerator(t *testing.T) {
	type args struct {
		min      int
		max      int
		maxRetry int
	}
	tests := []struct {
		name         string
		args         args
		times        int
		wantErr      bool
		maxRetry     bool
		excludePorts bool
	}{
		{
			name: "Should work - range 0",
			args: args{
				min: 1,
				max: 1,
			},
			times:        3,
			wantErr:      false,
			maxRetry:     false,
			excludePorts: false,
		},
		{
			name: "Should work - range 3",
			args: args{
				min: 1,
				max: 3,
			},
			times:        3,
			wantErr:      false,
			maxRetry:     false,
			excludePorts: false,
		},
		{
			name: "Should fail - max less than min",
			args: args{
				min: 1,
				max: 0,
			},
			times:        3,
			wantErr:      true,
			maxRetry:     false,
			excludePorts: false,
		},
		{
			name: "Should work - with memory",
			args: args{
				min: 1,
				max: 3,
			},
			times:        2,
			wantErr:      false,
			maxRetry:     false,
			excludePorts: true,
		},
		{
			name: "Should fail - all options saturated in the specified range",
			args: args{
				min: 1,
				max: 3,
			},
			times:        10,
			wantErr:      true,
			maxRetry:     false,
			excludePorts: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(tt.args.min, tt.args.max, tt.args.maxRetry, tt.excludePorts)

			// Errored, but don't wanted.
			if (err != nil) && !tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			// Errored, and wanted. Need to stop or ``.Generate` will be called
			// on `nil` causing `panic`.
			if (err != nil) && tt.wantErr {
				return
			}

			// Control flag.
			failed := false

			for i := 0; i < tt.times; i++ {
				_, err := r.Generate()
				if err != nil {
					failed = true
				}
			}

			if failed != tt.wantErr {
				t.Errorf("RandomPortGenerator() error = %v, wantErr %v", failed, tt.wantErr)

				return
			}
		})
	}
}
