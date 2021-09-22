// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package credential

import (
	"reflect"
	"testing"
)

func TestNewBasicAuthFromText(t *testing.T) {
	type args struct {
		credential string
	}
	tests := []struct {
		name    string
		args    args
		want    *BasicAuth
		wantErr bool
	}{
		{
			name:    "Should work - missing credential - nil",
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should work - missing credential - empty",
			args: args{
				credential: "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should work - credential required - only username",
			args: args{
				credential: "username",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should work - invalid credential - missing username",
			args: args{
				credential: ":password",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should work - invalid credential - missing password",
			args: args{
				credential: "username:",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should work - invalid credential - missing username, and password",
			args: args{
				credential: ":",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should work",
			args: args{
				credential: "username:password",
			},
			want: &BasicAuth{
				Username: "username",
				Password: "password",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBasicAuthFromText(tt.args.credential)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBasicAuthFromText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBasicAuthFromText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBasicAuth_ToBase64(t *testing.T) {
	type fields struct {
		Username string
		Password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Should word",
			fields: fields{
				Username: "user",
				Password: "pass",
			},
			want: "dXNlcjpwYXNz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bC := &BasicAuth{
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			if got := bC.ToBase64(); got != tt.want {
				t.Errorf("BasicAuth.ToBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}
