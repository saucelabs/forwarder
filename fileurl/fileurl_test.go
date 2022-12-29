// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package fileurl

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseFilePathOrURL(t *testing.T) {
	// Test cases use examples from RFC 8089.
	// See https://datatracker.ietf.org/doc/html/rfc8089
	tests := []struct {
		input string
		want  url.URL
	}{
		{
			input: "-",
			want:  url.URL{Scheme: "file", Path: "-"},
		},
		{
			input: "path/to/file",
			want:  url.URL{Scheme: "file", Path: "path/to/file"},
		},
		{
			input: "./path/to/file",
			want:  url.URL{Scheme: "file", Path: "./path/to/file"},
		},
		{
			input: "/path/to/file",
			want:  url.URL{Scheme: "file", Path: "/path/to/file"},
		},
		{
			input: "file:/path/to/file",
			want:  url.URL{Scheme: "file", Path: "/path/to/file"},
		},
		{
			input: "file:///path/to/file",
			want:  url.URL{Scheme: "file", Path: "/path/to/file"},
		},
		{
			input: "file://host.example.com/path/to/file",
			want: url.URL{
				Scheme: "file",
				Host:   "host.example.com",
				Path:   "/path/to/file",
			},
		},
		{
			input: "file:c:/path/to/file",
			want:  url.URL{Scheme: "file", Path: "c:/path/to/file"},
		},
		{
			input: "file:///c:/path/to/file",
			want:  url.URL{Scheme: "file", Path: "c:/path/to/file"},
		},
		{
			input: "file:///c|/path/to/file",
			want:  url.URL{Scheme: "file", Path: "c:/path/to/file"},
		},
		{
			input: "file:/c|/path/to/file",
			want:  url.URL{Scheme: "file", Path: "c:/path/to/file"},
		},
		{
			input: "file:c|/path/to/file",
			want:  url.URL{Scheme: "file", Path: "c:/path/to/file"},
		},
		{
			input: `\\host.example.com\Share\path\to\file.txt`,
			want: url.URL{
				Scheme: "file",
				Host:   "host.example.com",
				Path:   "/Share/path/to/file.txt",
			},
		},
		{
			input: "file:////host.example.com/path/to/file",
			want: url.URL{
				Scheme: "file",
				Host:   "host.example.com",
				Path:   "/path/to/file",
			},
		},
		{
			input: "file://///host.example.com/path/to/file",
			want: url.URL{
				Scheme: "file",
				Host:   "host.example.com",
				Path:   "/path/to/file",
			},
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseFilePathOrURL(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, *got, cmpopts.IgnoreFields(url.URL{}, "RawPath")); diff != "" {
				t.Errorf("ParseFilePathOrURL(%q) mismatch (-want +got):\n%s", tc.input, diff)
			}
			if got.EscapedPath() != got.Path {
				t.Errorf("ParseFilePathOrURL(%q) EscapedPath() = %q, want %q", tc.input, got.EscapedPath(), got.Path)
			}
		})
	}
}
