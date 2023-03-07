package e2e

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/gorilla/websocket"
)

var (
	proxy              = flag.String("proxy", "", "URL of the proxy to test against")
	httpbin            = flag.String("httpbin", "", "URL of the httpbin server to test against")
	maxWait            = flag.Duration("max-wait", 5*time.Second, "Maximum time to wait for the containers to become ready")
	insecureSkipVerify = flag.Bool("insecure-skip-verify", false, "Skip TLS certificate verification")
)

func init() {
	if os.Getenv("DEV") != "" {
		*proxy = "http://localhost:3128"
		*httpbin = "http://httpbin"
		*insecureSkipVerify = true
	}
}

// waitForServerReady checks the API server /readyz endpoint until it returns 200.
// It assumes that the server is running on port 10000.
func waitForServerReady(baseURL string) error {
	var client http.Client

	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	readyz := fmt.Sprintf("http://%s:10000/readyz", u.Hostname())

	req, err := http.NewRequest(http.MethodGet, readyz, http.NoBody)
	if err != nil {
		return err
	}

	const backoff = 200 * time.Millisecond

	var (
		resp *http.Response
		rerr error
	)
	for i := 0; i < int(*maxWait/backoff); i++ {
		resp, rerr = client.Do(req.Clone(context.Background()))

		if resp != nil && resp.StatusCode == http.StatusOK {
			return nil
		}

		time.Sleep(backoff)
	}
	if rerr != nil {
		return fmt.Errorf("%s not ready: %w", u.Hostname(), rerr)
	}

	return fmt.Errorf("%s not ready", u.Hostname())
}

func newTransport(t testing.TB) *http.Transport {
	t.Helper()

	if *proxy == "" {
		t.Fatal("proxy URL not set")
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: *insecureSkipVerify,
	}

	if *proxy == "" {
		t.Log("proxy not set, running without proxy")
	} else {
		proxyURL, err := url.Parse(*proxy)
		if err != nil {
			t.Fatal(err)
		}
		if ba := os.Getenv("FORWARDER_BASIC_AUTH"); ba != "" {
			u, p, _ := strings.Cut(ba, ":")
			proxyURL.User = url.UserPassword(u, p)
			t.Log("using basic auth for proxy", proxyURL)
		}
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	return tr
}

type client struct {
	tr *http.Transport
}

func (c client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.tr.RoundTrip(req)

	// There is a difference between sending HTTP and HTTPS requests.
	// For HTTPS client issues a CONNECT request to the proxy and then sends the original request.
	// In case the proxy responds with status code 4XX or 5XX to the CONNECT request, the client interprets it as URL error.
	//
	// This is to cover this case.
	if req.URL.Scheme == "https" && err != nil {
		for i := 400; i < 600; i++ {
			if err.Error() == http.StatusText(i) {
				return &http.Response{
					StatusCode: i,
					Status:     http.StatusText(i),
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header:     http.Header{},
					Body:       http.NoBody,
					Request:    req,
				}, nil
			}
		}
	}

	return resp, err
}

func Expect(t *testing.T, baseURL string, opts ...func(*httpexpect.Config)) *httpexpect.Expect {
	tr := newTransport(t)
	cfg := httpexpect.Config{
		BaseURL:  baseURL,
		Client:   client{tr: tr},
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
		WebsocketDialer: &websocket.Dialer{
			Proxy:           tr.Proxy,
			TLSClientConfig: tr.TLSClientConfig,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return httpexpect.WithConfig(cfg)
}

func ProxyNoAuth(config *httpexpect.Config) {
	tr := config.Client.(client).tr
	p := tr.Proxy
	tr.Proxy = func(req *http.Request) (u *url.URL, err error) {
		u, err = p(req)
		if u != nil {
			u.User = nil
		}
		return
	}
}
