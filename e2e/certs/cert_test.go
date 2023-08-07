//go:build manual

package certs_test

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/saucelabs/forwarder"
)

func TestCertificate(t *testing.T) {
	server := http.Server{
		Addr: "127.0.0.1:8443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello, world!"))
		}),
	}
	defer server.Close()

	go server.ListenAndServeTLS("httpbin.crt", "httpbin.key")

	tlsCfg := &tls.Config{
		ServerName: "httpbin",
	}
	cfg := forwarder.TLSClientConfig{
		CACertFiles: []string{
			"./ca.crt",
		},
	}
	cfg.ConfigureTLSConfig(tlsCfg)

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = tlsCfg

	req, err := http.NewRequest("GET", "https://"+server.Addr, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	res, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatal("unexpected status code:", res.StatusCode)
	}

	tr.CloseIdleConnections()
}
