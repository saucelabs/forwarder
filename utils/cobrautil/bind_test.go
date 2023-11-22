// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cobrautil

import (
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mmatczuk/anyflag"
	"github.com/spf13/cobra"
)

type testSliceStruct struct {
	Strings []string
	Ints    []int
	Bools   []bool
	IPs     []netip.Addr
}

func TestBindSlice(t *testing.T) {
	formats := []string{
		"yaml",
		"json",
		"toml",
	}

	for _, ext := range formats {
		t.Run(ext, func(t *testing.T) {
			cmd := &cobra.Command{}
			fs := cmd.Flags()

			var v testSliceStruct
			fs.String("config-file", "testdata/bind-slice."+ext, "")
			fs.StringSliceVar(&v.Strings, "strings", nil, "")
			fs.IntSliceVar(&v.Ints, "ints", nil, "")
			fs.BoolSliceVar(&v.Bools, "bools", nil, "")
			fs.Var(anyflag.NewSliceValue[netip.Addr](nil, &v.IPs, netip.ParseAddr), "ips", "")

			if err := BindAll(cmd, "TEST", "config-file"); err != nil {
				t.Fatal(err)
			}

			expected := testSliceStruct{
				Strings: []string{"a", "b", "c"},
				Ints:    []int{1, 2, 3},
				Bools:   []bool{true, false},
				IPs: []netip.Addr{
					netip.MustParseAddr("127.0.0.1"),
					netip.MustParseAddr("127.0.0.2"),
				},
			}

			ipcmp := cmp.Comparer(func(a, b netip.Addr) bool {
				return a.String() == b.String()
			})
			if diff := cmp.Diff(expected, v, ipcmp); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
