// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package pac

import "net"

func myIPAddress(ipv6 bool) (ips []net.IP) {
	ifces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for i := range ifces {
		if ifces[i].Flags&net.FlagUp != net.FlagUp {
			continue
		}
		addrs, err := ifces[i].Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, ok := addr.(*net.IPNet)
			if ok && ip.IP.IsGlobalUnicast() && (ipv6 || ip.IP.To4() != nil) {
				ips = append(ips, ip.IP)
			}
		}
	}
	return
}
