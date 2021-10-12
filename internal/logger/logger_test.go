// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package logger

import (
	"testing"
)

func TestSetup(t *testing.T) {
	type args struct {
		lvl       string
		fileLevel string
		filePath  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Should work",
			args: args{
				lvl:       "",
				fileLevel: "",
				filePath:  "",
			},
		},
		{
			name: "Should work",
			args: args{
				lvl:       infoLevel,
				fileLevel: infoLevel,
				filePath:  "-",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Setup(&Options{
				Level:     tt.args.lvl,
				FileLevel: tt.args.fileLevel,
				FilePath:  tt.args.filePath,
			})
			if l == nil {
				t.Errorf("Setup() expect %v to don't be nil", l)
			}

			if retrievedL := Get(); retrievedL == nil {
				t.Errorf("Get() expect %v to don't be nil", l)
			}

			// Should do nothing.
			l = Get()
			if l == nil {
				t.Errorf("Setup() expect %v to don't be nil", l)
			}
		})
	}
}
