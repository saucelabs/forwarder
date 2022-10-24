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

	"github.com/saucelabs/forwarder/log/stdlog"
	"github.com/saucelabs/forwarder/middleware"
)

const (
	defaultProxyHostname            = "127.0.0.1"
	defaultProxyScheme              = "http"
	upstreamProxyCredentialPassword = "p456"
	upstreamProxyCredentialUsername = "u456"
)

func TestHTTPProxyConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		config HTTPProxyConfig
		err    string
	}{
		{
			name: "both upstream and PAC are set",
			config: HTTPProxyConfig{
				UpstreamProxy: newProxyURL(80, upstreamProxyCredentialUsername, upstreamProxyCredentialPassword),
				PAC:           newProxyURL(80, "", ""),
			},
			err: "only one of upstream_proxy_uri or pac_uri can be set",
		},
		{
			name: "invalid upstream proxy ",
			config: HTTPProxyConfig{
				UpstreamProxy: &url.URL{},
			},
			err: "upstream_proxy_uri: invalid scheme",
		},
		{
			name: "invalid PAC server ",
			config: HTTPProxyConfig{
				PAC: &url.URL{},
			},
			err: "pac_uri: invalid scheme",
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err != nil {
				if tc.err == "" {
					t.Fatalf("expected success, got %q", err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("expected error to contain %q, got %q", tc.err, err)
				}
				return
			}
		})
	}
}

func TestHTTPProxySmoke(t *testing.T) { //nolint // FIXME cognitive complexity 88 of func `TestHTTPProxySmoke` is high (> 40) (gocognit); calculated cyclomatic complexity for function TestHTTPProxySmoke is 28, max is 10 (cyclop)
	tests := []struct {
		name    string
		config  HTTPProxyConfig
		wantPAC bool
	}{
		{
			name: "Local proxy",
			config: HTTPProxyConfig{
				ProxyLocalhost: true,
			},
		},
		{
			name: "Local proxy with site auth",
			config: HTTPProxyConfig{
				SiteCredentials: []string{},
				ProxyLocalhost:  true,
			},
		},
		{
			name: "Protected upstream proxy",
			config: HTTPProxyConfig{
				UpstreamProxy:  newProxyURL(0, upstreamProxyCredentialUsername, upstreamProxyCredentialPassword),
				ProxyLocalhost: true,
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
			targetServer := httpServerStub("body", targetCreds, stdlog.Default().Named("target"))
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

			// Upstream HTTPProxy.
			if tc.config.UpstreamProxy != nil {
				upstreamProxy, err := NewHTTPProxy(&HTTPProxyConfig{ProxyLocalhost: true}, nil, stdlog.Default().Named("upstream"))
				if err != nil {
					t.Fatalf("NewHTTPProxy() error=%v", err)
				}
				var h http.Handler = upstreamProxy
				if tc.config.UpstreamProxy.User != nil {
					p, _ := tc.config.UpstreamProxy.User.Password()
					h = middleware.NewProxyBasicAuth().Wrap(upstreamProxy, tc.config.UpstreamProxy.User.Username(), p)
				}

				usrv := httptest.NewServer(h)
				for usrv.Listener.Addr() == nil {
					time.Sleep(time.Millisecond)
				}
				defer usrv.Close()

				// Update the upstream proxy  with host and port.
				tc.config.UpstreamProxy.Host = usrv.Listener.Addr().String()
			}

			// Local proxy.
			localProxy, err := NewHTTPProxy(&tc.config, nil, stdlog.Default().Named("local"))
			if err != nil {
				t.Fatalf("NewHTTPProxy() error=%v", err)
			}
			lsrv := httptest.NewServer(localProxy)
			for lsrv.Listener.Addr() == nil {
				time.Sleep(time.Millisecond)
			}
			defer lsrv.Close()

			// Client's proxy settings.
			u, err := url.Parse(lsrv.URL)
			if err != nil {
				t.Fatalf("url.Parse() error=%v", err)
			}
			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(u),
				},
			}

			assertRequest(t, client, targetServerURL, http.StatusOK)
		})
	}
}

func TestHTTPProxyLocalhost(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()

	tests := []struct {
		name           string
		ProxyLocalhost bool
		StatusCode     int
	}{
		{
			name:           "enabled",
			ProxyLocalhost: true,
			StatusCode:     http.StatusOK,
		},
		{
			name:           "disabled",
			ProxyLocalhost: false,
			StatusCode:     http.StatusBadGateway,
		},
	}

	for i := range tests {
		tc := tests[i]

		t.Run(tc.name, func(t *testing.T) {
			// Start local proxy.
			localProxy, err := NewHTTPProxy(&HTTPProxyConfig{
				ProxyLocalhost: tc.ProxyLocalhost,
			}, nil, stdlog.Default().Named("local"))
			if err != nil {
				t.Fatalf("NewHTTPProxy() error = %v", err)
			}
			lsrv := httptest.NewServer(localProxy)
			for lsrv.Listener.Addr() == nil {
				time.Sleep(time.Millisecond)
			}
			defer lsrv.Close()

			// Client's proxy settings.
			u, err := url.Parse(lsrv.URL)
			if err != nil {
				t.Fatalf("url.Parse() error=%v", err)
			}
			t.Logf("Client is using %s as proxy", lsrv.URL)
			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(u),
				},
			}

			assertRequest(t, client, s.URL, tc.StatusCode)
		})
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
	for s.Listener.Addr() == nil {
		time.Sleep(time.Millisecond)
	}

	return s
}

func assertRequest(t *testing.T, client *http.Client, uri string, statusCode int) {
	t.Helper()

	u, err := url.ParseRequestURI(uri)
	if err != nil {
		t.Fatalf("Failed to parse : %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}

	if u.User != nil {
		password, _ := u.User.Password()

		request.SetBasicAuth(u.User.Username(), password)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Failed to execute request: %s", err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %s", err)
	}

	t.Logf("Response: %s", string(data))

	if response.StatusCode != statusCode {
		t.Fatalf("Expected status code to be %d, got %d", statusCode, response.StatusCode)
	}
}
