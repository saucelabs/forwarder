// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/saucelabs/randomness"
)

const (
	defaultProxyHostname            = "127.0.0.1"
	defaultProxyScheme              = "http"
	localProxyCredentialPassword    = "p123"
	localProxyCredentialUsername    = "u123"
	upstreamProxyCredentialPassword = "p456"
	upstreamProxyCredentialUsername = "u456"

	pacTemplate = `function FindProxyForURL(url, host) {
  if (
    dnsDomainIs(host, "intranet.domain.com") ||
    shExpMatch(host, "(*.abcdomain.com|abcdomain.com)")
  )
    return "DIRECT";

  return "PROXY 127.0.0.1:{{ .port }}; DIRECT";
}
`
)

func TestNewProxy(t *testing.T) { //nolint // FIXME cognitive complexity 88 of func `TestNewProxy` is high (> 40) (gocognit); calculated cyclomatic complexity for function TestNewProxy is 28, max is 10 (cyclop)
	//////
	// Randomness automates port allocation, ensuring no collision happens
	// between tests, and examples.
	//////

	r, err := randomness.New(10000, 20000, 100, true)
	if err != nil {
		t.Fatal("Failed to create randomness", err)
	}

	tests := []struct {
		name    string
		config  ProxyConfig
		wantPAC bool
	}{
		{
			name: "Local proxy",
			config: ProxyConfig{
				LocalProxyURI: newProxyURL(r.MustGenerate(), "", ""),
			},
		},
		{
			name: "Local proxy with site auth",
			config: ProxyConfig{
				LocalProxyURI:   newProxyURL(r.MustGenerate(), "", ""),
				SiteCredentials: []string{},
			},
		},
		{
			name: "Local proxy with DNS",
			config: ProxyConfig{
				DNSURIs:       []*url.URL{{Scheme: "udp", Host: "1.1.1.1:53"}},
				LocalProxyURI: newProxyURL(r.MustGenerate(), "", ""),
			},
		},
		{
			name: "Protected local proxy",
			config: ProxyConfig{
				LocalProxyURI: newProxyURL(r.MustGenerate(), localProxyCredentialUsername, localProxyCredentialPassword),
			},
		},
		{
			name: "Protected local proxy, and upstream proxy",
			config: ProxyConfig{
				LocalProxyURI:    newProxyURL(r.MustGenerate(), localProxyCredentialUsername, localProxyCredentialPassword),
				UpstreamProxyURI: newProxyURL(r.MustGenerate(), "", ""),
			},
		},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			targetCreds := ""
			if tc.config.SiteCredentials != nil {
				targetCreds = base64.
					StdEncoding.
					EncodeToString([]byte("user:pass"))
			}
			targetServer := httpServerStub("body", targetCreds, namedStdLogger("target"))
			defer func() {
				targetServer.Close()
			}()
			targetServerURL := targetServer.URL

			// Live test. Test calls to non-localhost. It matters because non-localhost
			// uses the proxy settings in the Transport, while localhost calls, bypass
			// it.
			if os.Getenv("FORWARDER_TEST_MODE") == "integration" {
				targetServerURL = "https://httpbin.org/status/200"
			}

			t.Logf("Target/end server @ %s", targetServerURL)

			if tc.config.SiteCredentials != nil {
				uri, err := url.Parse(targetServerURL)
				if err != nil {
					panic(err)
				}
				tc.config.SiteCredentials = append(tc.config.SiteCredentials, "user:pass@"+uri.Host)
			}

			//////
			// Local proxy.
			//
			// It's protected with Basic Auth. Upstream proxy will be automatically, and
			// dynamically setup via PAC, including credentials for proxies specified
			// in the PAC content.
			//////

			localProxy, err := NewProxy(tc.config, namedStdLogger("local"))
			if err != nil {
				t.Fatalf("NewProxy() error = %v", err)
			}
			go localProxy.MustRun()
			// Give enough time to start, and be ready.
			time.Sleep(1 * time.Second)

			//////
			// Upstream Proxy.
			//////

			if tc.config.UpstreamProxyURI != nil {
				upstreamProxy, err := NewProxy(ProxyConfig{LocalProxyURI: tc.config.UpstreamProxyURI}, namedStdLogger("upstream"))
				if err != nil {
					t.Fatalf("NewProxy() error = %v", err)
				}

				go upstreamProxy.MustRun()

				// Give enough time to start, and be ready.
				time.Sleep(1 * time.Second)
			}

			//////
			// Client.
			//////

			t.Logf("Client is using %s as proxy", tc.config.LocalProxyURI.Redacted())

			// Client's proxy settings.
			tr := &http.Transport{
				Proxy: http.ProxyURL(tc.config.LocalProxyURI),
			}

			client := &http.Client{
				Transport: tr,
			}

			statusCode, _, err := executeRequest(client, targetServerURL)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			if statusCode != http.StatusOK {
				t.Fatal("Expected status code to be OK")
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	//////
	// Target/end server.
	//////

	testServer := httpServerStub("body", "", nopLogger{})

	defer func() { testServer.Close() }()

	//////
	// Proxy.
	//////

	r, err := randomness.New(30000, 40000, 100, true)
	if err != nil {
		b.Fatal("Failed to create proxy", err)
	}

	localProxyURI := newProxyURL(r.MustGenerate(), "", "")

	proxy, err := NewProxy(ProxyConfig{LocalProxyURI: localProxyURI}, nopLogger{})
	if err != nil {
		b.Fatal("Failed to create proxy.", err)
	}

	go proxy.MustRun()

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	//////
	// Client.
	//////

	// Client's proxy settings.
	tr := &http.Transport{
		Proxy: http.ProxyURL(localProxyURI),
	}

	client := &http.Client{
		Transport: tr,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, err := executeRequest(client, testServer.URL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// newProxyURL returns a URL with the given scheme, host, port and path.
func newProxyURL(port int64, username, password string) *url.URL {
	u := &url.URL{
		Scheme: defaultProxyScheme,
		Host:   fmt.Sprintf("%s:%d", defaultProxyHostname, port),
	}

	if username != "" && password != "" {
		u.User = url.UserPassword(username, password)
	}

	return u
}

func httpServerStub(body, encodedCredential string, log Logger) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if encodedCredential != "" {
			ok := strings.Contains(req.Header.Get("Authorization"), encodedCredential)

			log.Debugf("Incoming request. This server (%s) is protected authorized=%v", req.Host, ok)

			if !ok {
				http.Error(res, http.StatusText(http.StatusForbidden), http.StatusForbidden)

				return
			}
		}

		res.WriteHeader(http.StatusOK)

		if _, err := res.Write([]byte(body)); err != nil {
			http.Error(res, err.Error(), http.StatusForbidden)

			return
		}
	}))

	// Wait for the server to start.
	time.Sleep(1 * time.Second)

	return s
}

func executeRequest(client *http.Client, uri string) (statusCode int, body string, err error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return 0, "", fmt.Errorf("Failed to parse URI: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return 0, "", fmt.Errorf("Failed to create request: %w", err)
	}

	if u.User != nil {
		password, _ := u.User.Password()

		request.SetBasicAuth(u.User.Username(), password)
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, "", fmt.Errorf("Failed to execute request: %w", err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, "", fmt.Errorf("Failed to read body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("Failed request, non-2xx code (%d): %s", response.StatusCode, data)
	}

	return response.StatusCode, string(data), nil
}
