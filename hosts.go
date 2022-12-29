// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import (
	"bytes"
	"net"
	"os"

	hostsfile "github.com/kevinburke/hostsfile/lib"
)

// ReadHostsFile reads the /etc/hosts file and returns a map of hostnames to IP addresses.
// If the file does not exist, a nil map is returned.
func ReadHostsFile() (map[string]net.IP, error) {
	b, err := os.ReadFile(hostsfile.Location)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return nil, err
	}

	lines, err := hostsfile.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	m := make(map[string]net.IP, len(lines.Records()))
	for _, r := range lines.Records() {
		for h := range r.Hostnames {
			m[h] = r.IpAddress.IP
		}
	}

	return m, nil
}
