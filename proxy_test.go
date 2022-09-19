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
	defaultProxyScheme              = HTTP
	localProxyCredentialPassword    = "p123"
	localProxyCredentialUsername    = "u123"
	upstreamProxyCredentialPassword = "p456"
	upstreamProxyCredentialUsername = "u456"
	wrongCredentialPassword         = "wrongPassword"
	wrongCredentialUsername         = "wrongUsername"

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

func TestProxyConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		config ProxyConfig
		err    string
	}{
		{
			name: "Valid",
			config: ProxyConfig{
				LocalProxyURI: "http://foobar:1234",
			},
		},
		{
			name: "Both upstream and PAC are set",
			config: ProxyConfig{
				LocalProxyURI:    newProxyURL(80, localProxyCredentialUsername, localProxyCredentialPassword).String(),
				UpstreamProxyURI: newProxyURL(80, upstreamProxyCredentialUsername, upstreamProxyCredentialPassword).String(),
				PACURI:           newProxyURL(80, "", "").String(),
			},
			err: "excluded_with",
		},
		{
			name: "Missing local proxy URI",
			err:  "required",
		},
		{
			name: "Invalid local proxy URI",
			config: ProxyConfig{
				LocalProxyURI: "foo",
			},
			err: "proxyURI",
		},
		{
			name: "Invalid upstream proxy URI",
			config: ProxyConfig{
				LocalProxyURI:    newProxyURL(80, "", "").String(),
				UpstreamProxyURI: "foo",
			},
			err: "proxyURI",
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err != nil && tc.err == "" {
				t.Fatalf("Expected no error, got %s", err)
			}

			if err == nil && tc.err != "" {
				t.Fatal("Expected error, got none")
			}

			if err != nil && tc.err != "" && !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("Expected error to contain %s, got %s", tc.err, err)
			}
		})
	}
}

func TestNewProxy(t *testing.T) { //nolint // FIXME cognitive complexity 88 of func `TestNewProxy` is high (> 40) (gocognit); calculated cyclomatic complexity for function TestNewProxy is 28, max is 10 (cyclop)
	//////
	// Randomness automates port allocation, ensuring no collision happens
	// between tests, and examples.
	//////

	r, err := randomness.New(10000, 20000, 100, true)
	if err != nil {
		t.Fatal("Failed to create randomness", err)
	}

	type args struct {
		dnsURIs               []string
		localProxyURI         *url.URL
		upstreamProxyURI      *url.URL
		pacURI                *url.URL
		pacProxiesCredentials []string
		siteCredentials       []string
	}
	tests := []struct {
		name             string
		args             args
		preFunc          func()
		postFunc         func()
		preUpstreamFunc  func()
		postUpstreamFunc func()
		wantPAC          bool
	}{
		{
			name: "Local proxy",
			args: args{
				localProxyURI: newProxyURL(r.MustGenerate(), "", ""),
			},
		},
		{
			name: "Local proxy with site auth",
			args: args{
				localProxyURI:   newProxyURL(r.MustGenerate(), "", ""),
				siteCredentials: []string{},
			},
		},
		{
			name: "Local proxy with DNS",
			args: args{
				dnsURIs:       []string{"udp://8.8.8.8:53"},
				localProxyURI: newProxyURL(r.MustGenerate(), "", ""),
			},
		},
		{
			name: "Protected local proxy",
			args: args{
				localProxyURI: newProxyURL(r.MustGenerate(), localProxyCredentialUsername, localProxyCredentialPassword),
			},
		},
		{
			name: "Protected local proxy, and upstream proxy",
			args: args{
				localProxyURI:    newProxyURL(r.MustGenerate(), localProxyCredentialUsername, localProxyCredentialPassword),
				upstreamProxyURI: newProxyURL(r.MustGenerate(), "", ""),
			},
		},
		{
			name: "Protected local proxy, and protected upstream proxy",
			args: args{
				localProxyURI:    newProxyURL(r.MustGenerate(), wrongCredentialUsername, wrongCredentialPassword),
				upstreamProxyURI: newProxyURL(r.MustGenerate(), wrongCredentialUsername, wrongCredentialPassword),
			},
			preFunc: func() {
				// Local proxy.
				os.Setenv("FORWARDER_LOCALPROXY_AUTH", url.UserPassword(
					localProxyCredentialUsername,
					localProxyCredentialPassword,
				).String())

				// Upstream proxy.
				os.Setenv("FORWARDER_UPSTREAMPROXY_AUTH", url.UserPassword(
					upstreamProxyCredentialUsername,
					upstreamProxyCredentialPassword,
				).String())
			},
			postFunc: func() {
				os.Unsetenv("FORWARDER_LOCALPROXY_AUTH")

				os.Unsetenv("FORWARDER_UPSTREAMPROXY_AUTH")
			},
			preUpstreamFunc: func() {
				// Local proxy.
				os.Setenv("FORWARDER_LOCALPROXY_AUTH", url.UserPassword(
					upstreamProxyCredentialUsername,
					upstreamProxyCredentialPassword,
				).String())
			},
			postUpstreamFunc: func() {
				os.Unsetenv("FORWARDER_LOCALPROXY_AUTH")
			},
		},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if tc.preFunc != nil {
				tc.preFunc()
			}

			//////
			// Target/end server.
			//////

			targetCreds := ""
			if tc.args.siteCredentials != nil {
				targetCreds = base64.
					StdEncoding.
					EncodeToString([]byte("user:pass"))
			}
			targetServer := httpServerStub("body", targetCreds, namedStdLogger("target"))

			defer func() { targetServer.Close() }()

			targetServerURL := targetServer.URL

			// Live test. Test calls to non-localhost. It matters because non-localhost
			// uses the proxy settings in the Transport, while localhost calls, bypass
			// it.
			var dnsURIs []string

			if os.Getenv("FORWARDER_TEST_MODE") == "integration" {
				targetServerURL = "https://httpbin.org/status/200"

				if tc.args.dnsURIs != nil {
					dnsURIs = tc.args.dnsURIs
				}
			}

			t.Logf("Target/end server @ %s", targetServerURL)

			//////
			// Set proxy URIs - if any.
			//////

			localProxyURI := ""
			if tc.args.localProxyURI != nil {
				localProxyURI = tc.args.localProxyURI.String()
			}

			upstreamProxyURI := ""
			if tc.args.upstreamProxyURI != nil {
				upstreamProxyURI = tc.args.upstreamProxyURI.String()
			}

			pacURI := ""
			if tc.args.pacURI != nil {
				pacURI = tc.args.pacURI.String()
			}

			var siteCredentials []string
			if tc.args.siteCredentials != nil {
				uri, err := url.Parse(targetServerURL)
				if err != nil {
					panic(err)
				}

				siteCredentials = append(siteCredentials, "user:pass@"+uri.Host)
			}

			//////
			// Local proxy.
			//
			// It's protected with Basic Auth. Upstream proxy will be automatically, and
			// dynamically setup via PAC, including credentials for proxies specified
			// in the PAC content.
			//////

			c := ProxyConfig{
				LocalProxyURI:         localProxyURI,
				UpstreamProxyURI:      upstreamProxyURI,
				PACURI:                pacURI,
				PACProxiesCredentials: tc.args.pacProxiesCredentials,
				DNSURIs:               dnsURIs,
				SiteCredentials:       siteCredentials,
			}
			localProxy, err := NewProxy(c, namedStdLogger("local"))
			if err != nil {
				t.Fatalf("NewProxy() error = %v", err)
			}

			// Both local `localProxy.LocalProxyURI` and `localProxy.UpstreamProxyURI`
			// changed, if `FORWARDER_LOCALPROXY_AUTH` or `FORWARDER_UPSTREAMPROXY_AUTH`
			// were set. This updates test vars.
			if localProxyURI != localProxy.Config().LocalProxyURI {
				localProxyURI = localProxy.Config().LocalProxyURI

				lPURI, err := url.ParseRequestURI(localProxyURI)
				if err != nil {
					t.Fatal("Failed to ParseRequestURI(localProxyURI)")
				}

				tc.args.localProxyURI = lPURI
			}

			if upstreamProxyURI != localProxy.Config().UpstreamProxyURI {
				upstreamProxyURI = localProxy.Config().UpstreamProxyURI

				lUURI, err := url.ParseRequestURI(upstreamProxyURI)
				if err != nil {
					t.Fatal("Failed to ParseRequestURI(upstreamProxyURI)")
				}

				tc.args.upstreamProxyURI = lUURI
			}

			go localProxy.MustRun()

			// Give enough time to start, and be ready.
			time.Sleep(1 * time.Second)

			//////
			// Upstream Proxy.
			//////

			if tc.preUpstreamFunc != nil {
				tc.preUpstreamFunc()
			}

			if upstreamProxyURI != "" {
				upstreamProxy, err := NewProxy(ProxyConfig{LocalProxyURI: upstreamProxyURI}, namedStdLogger("upstream"))
				if err != nil {
					t.Fatalf("NewProxy() error = %v", err)
				}

				go upstreamProxy.MustRun()

				// Give enough time to start, and be ready.
				time.Sleep(1 * time.Second)
			}

			if tc.postUpstreamFunc != nil {
				tc.postUpstreamFunc()
			}

			//////
			// Client.
			//////

			t.Logf("Client is using %s as proxy", tc.args.localProxyURI.Redacted())

			// Client's proxy settings.
			tr := &http.Transport{
				Proxy: http.ProxyURL(tc.args.localProxyURI),
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

			if tc.postFunc != nil {
				tc.postFunc()
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

	proxy, err := NewProxy(ProxyConfig{LocalProxyURI: localProxyURI.String()}, nopLogger{})
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
