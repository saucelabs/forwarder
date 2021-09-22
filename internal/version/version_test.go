// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package version

import (
	"strings"
	"testing"
)

func Test_setupVersion(t *testing.T) {
	type args struct {
		buildCommit  string
		buildTime    string
		buildVersion string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Should work",
			args: args{
				buildCommit:  "1223423321234sdf",
				buildTime:    "2021-09-21T12:49:39-07:00",
				buildVersion: "v0.0.1",
			},
			want: []string{"v0.0.1", "1223423321234sdf"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildCommit = tt.args.buildCommit
			buildTime = tt.args.buildTime
			buildVersion = tt.args.buildVersion

			got := Get()

			control := true

			for _, w := range tt.want {
				if !strings.Contains(got.String(), w) {
					control = false
				}
			}

			if !control {
				t.Errorf("setupVersion() got = %v, expected to contain %v", got.String(), strings.Join(tt.want, ", "))

				return
			}
		})
	}
}
