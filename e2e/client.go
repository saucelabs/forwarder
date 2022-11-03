package e2e

import (
	"flag"
	"net/http"
	"net/url"
	"testing"

	"github.com/saucelabs/forwarder"
)

func newHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	if !flag.Parsed() {
		flag.Parse()
	}

	if *proxy == "" {
		t.Fatal("proxy URL not set")
	}

	cfg := forwarder.DefaultHTTPTransportConfig()
	cfg.InsecureSkipVerify = *insecureSkipVerify

	rt := forwarder.NewHTTPTransport(cfg, nil)
	if *proxy == "" {
		t.Log("proxy not set, running without proxy")
	} else {
		proxyURL, err := url.Parse(*proxy)
		if err != nil {
			t.Fatal(err)
		}
		rt.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: rt,
	}
}
