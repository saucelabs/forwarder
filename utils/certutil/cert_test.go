// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package certutil

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRSASelfSignedCertGen(t *testing.T) {
	cert, err := RSASelfSignedCert().Gen()
	if err != nil {
		t.Fatalf("RSASelfSignedCert.Gen() error %s", err)
	}
	s := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	s.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	defer s.Close()
	s.StartTLS()

	c := http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := c.Get(s.URL)
	if err != nil {
		t.Fatalf("http.Get() error %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("http.Get() status code %d", resp.StatusCode)
	}
}
