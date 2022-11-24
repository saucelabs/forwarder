package e2e

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

var (
	proxy              = flag.String("proxy", "", "URL of the proxy to test against")
	httpbin            = flag.String("httpbin", "", "URL of the httpbin server to test against")
	insecureSkipVerify = flag.Bool("insecure-skip-verify", false, "Skip TLS certificate verification")
)

func init() {
	if os.Getenv("DEV") != "" {
		*proxy = "https://localhost:3128"
		*httpbin = "https://httpbin"
		*insecureSkipVerify = true
	}
}

func newHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	if *proxy == "" {
		t.Fatal("proxy URL not set")
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: *insecureSkipVerify}

	if *proxy == "" {
		t.Log("proxy not set, running without proxy")
	} else {
		proxyURL, err := url.Parse(*proxy)
		if err != nil {
			t.Fatal(err)
		}
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	return &http.Client{
		Transport: tr,
	}
}

func expect(t *testing.T, baseURL string, opts ...func(*httpexpect.Config)) *httpexpect.Expect {
	cfg := httpexpect.Config{
		BaseURL:  baseURL,
		Client:   newHTTPClient(t),
		Reporter: t,
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return httpexpect.WithConfig(cfg)
}

func expectError(t *testing.T, client *http.Client, method, url string, body io.Reader, ck func(err error)) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err == nil {
		t.Fatal("expected error")
	}
	ck(err)
}

func errorMatches(t *testing.T, want string) func(err error) {
	t.Helper()
	return func(err error) {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected error %q, got %q", want, err.Error())
		}
	}
}
