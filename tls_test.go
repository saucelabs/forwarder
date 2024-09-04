// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"crypto/tls"
	"io"
	"net"
	"testing"

	"github.com/saucelabs/forwarder/utils/certutil"
	"golang.org/x/net/netutil"
)

func TestTLSClientConfigInsecure(t *testing.T) {
	t.Parallel()

	var tlsCfg tls.Config
	cc := TLSClientConfig{
		Insecure: true,
	}
	cc.ConfigureTLSConfig(&tlsCfg)

	for _, c := range tls.InsecureCipherSuites() {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			if c.Name == "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA" {
				t.Skip("RC4 is disabled by default")
			}
			if c.Name == "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256" {
				t.Skip("CBC_SHA256 is disabled by default")
			}

			l, err := tls.Listen("tcp", "localhost:0", tls12ServerTLSConfig(t, c.ID))
			if err != nil {
				t.Fatal(err)
			}
			l = netutil.LimitListener(l, 1)

			go func() {
				conn, err := l.Accept()
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				io.Copy(conn, conn)
			}()

			conn, err := net.Dial("tcp", l.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()

			tconn := tls.Client(conn, &tlsCfg)
			if err := tconn.Handshake(); err != nil {
				t.Fatal(err)
			}
			defer tconn.Close()
		})
	}
}

func tls12ServerTLSConfig(t *testing.T, cipherSuite uint16) *tls.Config {
	t.Helper()

	tlsCfg := tls.Config{
		MaxVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{cipherSuite},
	}

	ssc := certutil.RSASelfSignedCert()
	cert, err := ssc.Gen()
	if err != nil {
		t.Fatal(err)
	}
	tlsCfg.Certificates = []tls.Certificate{
		cert,
	}
	return &tlsCfg
}
