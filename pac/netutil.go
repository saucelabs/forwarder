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
