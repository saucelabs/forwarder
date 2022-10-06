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

func TestProxyConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		config ProxyConfig
		err    string
	}{
		{
			name: "both upstream and PAC are set",
			config: ProxyConfig{
				UpstreamProxyURI: newProxyURL(80, upstreamProxyCredentialUsername, upstreamProxyCredentialPassword),
				PACURI:           newProxyURL(80, "", ""),
			},
			err: "only one of upstream_proxy_uri or pac_uri can be set",
		},
		{
			name: "invalid upstream proxy URI",
			config: ProxyConfig{
				UpstreamProxyURI: &url.URL{},
			},
			err: "upstream_proxy_uri: invalid scheme",
		},
		{
			name: "invalid PAC server URI",
			config: ProxyConfig{
				PACURI: &url.URL{},
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

func TestNewProxy(t *testing.T) { //nolint // FIXME cognitive complexity 88 of func `TestNewProxy` is high (> 40) (gocognit); calculated cyclomatic complexity for function TestNewProxy is 28, max is 10 (cyclop)
	tests := []struct {
		name    string
		config  ProxyConfig
		wantPAC bool
	}{
		{
			name: "Local proxy",
			config: ProxyConfig{
				ProxyLocalhost: true,
			},
		},
		{
			name: "Local proxy with site auth",
			config: ProxyConfig{
				SiteCredentials: []string{},
				ProxyLocalhost:  true,
			},
		},
		{
			name: "Protected local proxy",
			config: ProxyConfig{
				BasicAuth:      url.UserPassword(localProxyCredentialUsername, localProxyCredentialPassword),
				ProxyLocalhost: true,
			},
		},
		{
			name: "Protected local proxy, and upstream proxy",
			config: ProxyConfig{
				BasicAuth:        url.UserPassword(localProxyCredentialUsername, localProxyCredentialPassword),
				UpstreamProxyURI: newProxyURL(0, "", ""),
				ProxyLocalhost:   true,
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

			// Upstream Proxy.
			if tc.config.UpstreamProxyURI != nil {
				upstreamProxy, err := NewProxy(&ProxyConfig{BasicAuth: tc.config.UpstreamProxyURI.User, ProxyLocalhost: true}, nil, namedStdLogger("upstream"))
				if err != nil {
					t.Fatalf("NewProxy() error=%v", err)
				}
				usrv := httptest.NewServer(upstreamProxy)
				for usrv.Listener.Addr() == nil {
					time.Sleep(time.Millisecond)
				}
				defer usrv.Close()

				// Update the upstream proxy URI with host and port.
				tc.config.UpstreamProxyURI.Host = usrv.Listener.Addr().String()
			}

			// Local proxy.
			localProxy, err := NewProxy(&tc.config, nil, namedStdLogger("local"))
			if err != nil {
				t.Fatalf("NewProxy() error=%v", err)
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
			u.User = tc.config.BasicAuth
			t.Logf("Client is using %s as proxy", u)
			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(u),
				},
			}

			if _, err := assertRequest(client, targetServerURL, http.StatusOK); err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
		})
	}
}

func TestProxyLocalhost(t *testing.T) {
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
			localProxy, err := NewProxy(&ProxyConfig{
				ProxyLocalhost: tc.ProxyLocalhost,
			}, nil, namedStdLogger("local"))
			if err != nil {
				t.Fatalf("NewProxy() error = %v", err)
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

			// Make request to localhost.
			if _, err := assertRequest(client, s.URL, tc.StatusCode); err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
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

	// Wait for the server to start.
	time.Sleep(1 * time.Second)

	return s
}

func assertRequest(client *http.Client, uri string, statusCode int) (body string, err error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", fmt.Errorf("Failed to parse URI: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return "", fmt.Errorf("Failed to create request: %w", err)
	}

	if u.User != nil {
		password, _ := u.User.Password()

		request.SetBasicAuth(u.User.Username(), password)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("Failed to execute request: %w", err)
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read body: %w", err)
	}

	if response.StatusCode != statusCode {
		return "", fmt.Errorf("Expected status code to be %d, got %d", statusCode, response.StatusCode)
	}

	return string(data), nil
}
