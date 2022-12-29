// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.
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
