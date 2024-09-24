//go:build ignore
package main

import (
	"net"
	"os"

	"github.com/saucelabs/forwarder/proxyproto"
)

func main() {
	h := proxyproto.Header{
		Version:           2,
		Command:           proxyproto.PROXY,
		TransportProtocol: proxyproto.TCPv4,
		SourceAddr:        &net.TCPAddr{
			IP:   net.ParseIP("1.1.1.1"),
			Port: 1000,
		},
		DestinationAddr:   &net.TCPAddr{
			IP:   net.ParseIP("2.2.2.2"),
			Port: 2000,
		},
	}

	f, err := os.Create("v2.bin")
	if err != nil {
		panic(err)
	}

	h.WriteTo(f)

	if err := f.Close(); err != nil {
		panic(err)
	}
}
