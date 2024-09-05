// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hostsfile

import (
	"io"
	"os"
	"sort"

	hostsfile "github.com/kevinburke/hostsfile/lib"
	"golang.org/x/exp/maps"
)

func LocalhostAliases() ([]string, error) {
	f, err := os.Open(hostsfile.Location)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return readLocalhostAliases(f)
}

func readLocalhostAliases(r io.Reader) ([]string, error) {
	hf, err := hostsfile.Decode(r)
	if err != nil {
		return nil, err
	}

	aliases := make(map[string]struct{})
	for _, r := range hf.Records() {
		if r.IpAddress.IP.IsLoopback() {
			for a := range r.Hostnames {
				aliases[a] = struct{}{}
			}
		}
	}

	v := maps.Keys(aliases)
	sort.Strings(v)
	return v, nil
}
