// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package pac

import (
	"context"
	"net"

	"github.com/dop251/goja"
)

// Handler for "dnsResolve(host)".
// Resolves the given DNS hostname into an IP address, and returns it in the dot-separated format as a string.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file#dnsresolve
func (pr *ProxyResolver) dnsResolve(call goja.FunctionCall) goja.Value {
	host, ok := asString(call.Argument(0))
	if !ok {
		return goja.Undefined()
	}

	lookupIP := pr.config.testingLookupIP
	if lookupIP == nil {
		lookupIP = pr.resolver.LookupIP
	}
	ips, err := lookupIP(context.Background(), "ip4", host)
	if err != nil {
		return goja.Null()
	}

	return pr.vm.ToValue(ips[0].String())
}

// Handler for "myIpAddress()".
// Returns the machine IP address as a string in the dot-separated integer format.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file#myipaddress
func (pr *ProxyResolver) myIPAddress(call goja.FunctionCall) goja.Value {
	var ips []net.IP
	if pr.config.testingMyIPAddress != nil {
		ips = pr.config.testingMyIPAddress
	} else {
		ips = myIPAddress(false)
	}

	if len(ips) == 0 {
		return pr.vm.ToValue("127.0.0.1")
	}

	return pr.vm.ToValue(ips[0].String())
}
