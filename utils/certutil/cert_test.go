// Copyright 2023 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !windows

package certutil

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRSASelfSignedCertGen(t *testing.T) {
	c := RSASelfSignedCert()
	c.Hosts = []string{"127.0.0.1"}

	cert, err := c.Gen()
	if err != nil {
		t.Fatalf("RSASelfSignedCert.Gen() error %s", err)
	}
	testCert(t, &cert)
}

func TestECDSASelfSignedCertGen(t *testing.T) {
	c := ECDSASelfSignedCert()
	c.Hosts = []string{"127.0.0.1"}

	cert, err := c.Gen()
	if err != nil {
		t.Fatalf("ECDSASelfSignedCert.Gen() error %s", err)
	}
	testCert(t, &cert)
}

func testCert(t *testing.T, cert *tls.Certificate) { //nolint:thelper // this is not a test helper
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	s.TLS = &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	defer s.Close()
	s.StartTLS()

	cacert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("x509.ParseCertificate() error %s", err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(cacert)

	c := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{
		RootCAs: pool,
	}}}
	resp, err := c.Get(s.URL)
	if err != nil {
		t.Fatalf("http.Get() error %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("http.Get() status code %d", resp.StatusCode)
	}
}
