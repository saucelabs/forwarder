// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/internal/logger"
	"github.com/saucelabs/randomness"
	"github.com/saucelabs/sypl/fields"
	"github.com/saucelabs/sypl/level"
	"github.com/saucelabs/sypl/options"
)

const (
	// Change to `Trace` for debugging, and demonstration purposes.
	defaultProxyLoggingLevel = "none"

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

var loggingOptions = &LoggingOptions{
	Level: defaultProxyLoggingLevel, // NOTE: Set it to `trace` to debug problems.
}

//////
// Helpers
//////

// Builds and URI and returns as string.
func URIBuilder(hostname string, port int64, username, password string) *url.URL {
	u := &url.URL{
		Scheme: defaultProxyScheme,
		Host:   fmt.Sprintf("%s:%d", hostname, port),
	}

	if username != "" && password != "" {
		u.User = url.UserPassword(username, password)
	}

	return u
}

// URLUserStripper removes `User` (credential) information from the URL.
func URLUserStripper(u *url.URL) *url.URL {
	lU := *u
	lU.User = nil

	return &lU
}

// Creates a mocked HTTP server. Don't forget to defer close it!
//nolint:unparam
func createMockedHTTPServer(statusCode int, body, encodedCredential string) *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if encodedCredential != "" {
			ok := strings.Contains(req.Header.Get("Authorization"), encodedCredential)

			logger.Get().PrintlnfWithOptions(&options.Options{
				Fields: fields.Fields{
					"authorized": ok,
				},
			}, level.Trace, "Incoming request. This server (%s) is protected", req.Host)

			if !ok {
				http.Error(res, http.StatusText(http.StatusForbidden), http.StatusForbidden)

				return
			}
		}

		res.WriteHeader(statusCode)

		if _, err := res.Write([]byte(body)); err != nil {
			http.Error(res, err.Error(), http.StatusForbidden)

			return
		}
	}))

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	return testServer
}

func executeRequest(client *http.Client, uri string) (int, string, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return 0, "", fmt.Errorf("Failed to parse URI: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, "", fmt.Errorf("Failed to read body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("Failed request, non-2xx code (%d): %s", response.StatusCode, data)
	}

	return response.StatusCode, string(data), nil
}

//////
// Tests
//////

//nolint:maintidx
func TestNew(t *testing.T) {
	//////
	// Randomness automates port allocation, ensuring no collision happens
	// between tests, and examples.
	//////

	r, err := randomness.New(10000, 20000, 100, true)
	if err != nil {
		log.Fatalln("Failed to create randomness.", err)
	}

	type args struct {
		dnsURIs               []string
		localProxyURI         *url.URL
		upstreamProxyURI      *url.URL
		pacURI                *url.URL
		pacProxiesCredentials []string
		loggingOptions        *LoggingOptions
	}
	tests := []struct {
		name             string
		args             args
		preFunc          func()
		postFunc         func()
		preUpstreamFunc  func()
		postUpstreamFunc func()
		wantPAC          bool
		wantErr          bool
		wantErrType      error
	}{
		{
			name: "Should work - local proxy",
			args: args{
				localProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					"",
					"",
				),
				loggingOptions: loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should work - local proxy - with DNS",
			args: args{
				dnsURIs: []string{"udp://8.8.8.8:53"},
				localProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					"",
					"",
				),
				loggingOptions: loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should work - protected local proxy",
			args: args{
				localProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					localProxyCredentialUsername,
					localProxyCredentialPassword,
				),
				loggingOptions: loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should work - protected local proxy, and upstream proxy",
			args: args{
				localProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					localProxyCredentialUsername,
					localProxyCredentialPassword,
				),
				upstreamProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					"",
					"",
				),
				loggingOptions: loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should work - protected local proxy, and protected upstream proxy",
			args: args{
				localProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					wrongCredentialUsername,
					wrongCredentialPassword,
				),
				upstreamProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					wrongCredentialUsername,
					wrongCredentialPassword,
				),
				loggingOptions: loggingOptions,
			},
			preFunc: func() {
				// Local proxy.
				os.Setenv("FORWARDER_LOCALPROXY_CREDENTIAL", url.UserPassword(
					localProxyCredentialUsername,
					localProxyCredentialPassword,
				).String())

				// Upstream proxy.
				os.Setenv("FORWARDER_UPSTREAMPROXY_CREDENTIAL", url.UserPassword(
					upstreamProxyCredentialUsername,
					upstreamProxyCredentialPassword,
				).String())
			},
			postFunc: func() {
				os.Unsetenv("FORWARDER_LOCALPROXY_CREDENTIAL")

				os.Unsetenv("FORWARDER_UPSTREAMPROXY_CREDENTIAL")
			},
			preUpstreamFunc: func() {
				// Local proxy.
				os.Setenv("FORWARDER_LOCALPROXY_CREDENTIAL", url.UserPassword(
					upstreamProxyCredentialUsername,
					upstreamProxyCredentialPassword,
				).String())
			},
			postUpstreamFunc: func() {
				os.Unsetenv("FORWARDER_LOCALPROXY_CREDENTIAL")
			},
			wantErr: false,
		},
		{
			name: "Should fail - both upstream and PAC are set",
			args: args{
				localProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					localProxyCredentialUsername,
					localProxyCredentialPassword,
				),
				upstreamProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					upstreamProxyCredentialUsername,
					upstreamProxyCredentialPassword,
				),
				pacURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					"",
					"",
				),
				loggingOptions: loggingOptions,
			},
			wantErr:     true,
			wantErrType: ErrInvalidOrParentOrPac,
		},
		{
			name: "Should fail - missing local proxy URI",
			args: args{
				loggingOptions: loggingOptions,
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyParams,
		},
		{
			name: "Should fail - invalid local proxy URI",
			args: args{
				localProxyURI:  URIBuilder("", 0, "", ""),
				loggingOptions: loggingOptions,
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyParams,
		},
		{
			name: "Should fail - invalid upstream proxy URI",
			args: args{
				localProxyURI:    URIBuilder(defaultProxyHostname, r.MustGenerate(), "", ""),
				upstreamProxyURI: URIBuilder("", 0, "", ""),
				loggingOptions:   loggingOptions,
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyParams,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preFunc != nil {
				tt.preFunc()
			}

			// Re-set the value allowing per-test case setting.
			l := logger.Setup(tt.args.loggingOptions)

			//////
			// Target/end server.
			//////

			targetServer := createMockedHTTPServer(http.StatusOK, "body", "")

			defer func() { targetServer.Close() }()

			targetServerURL := targetServer.URL

			// Live test. Test calls to non-localhost. It matters because non-localhost
			// uses the proxy settings in the Transport, while localhost calls, bypass
			// it.
			var dnsURIs []string

			if os.Getenv("FORWARDER_TEST_MODE") == "integration" {
				targetServerURL = "https://httpbin.org/status/200"

				if tt.args.dnsURIs != nil {
					dnsURIs = tt.args.dnsURIs
				}
			}

			l.Debuglnf("Target/end server @ %s", targetServerURL)

			//////
			// Set proxy URIs - if any.
			//////

			localProxyURI := ""
			if tt.args.localProxyURI != nil {
				localProxyURI = tt.args.localProxyURI.String()
			}

			upstreamProxyURI := ""
			if tt.args.upstreamProxyURI != nil {
				upstreamProxyURI = tt.args.upstreamProxyURI.String()
			}

			pacURI := ""
			if tt.args.pacURI != nil {
				pacURI = tt.args.pacURI.String()
			}

			//////
			// Local proxy.
			//
			// It's protected with Basic Auth. Upstream proxy will be automatically, and
			// dynamically setup via PAC, including credentials for proxies specified
			// in the PAC content.
			//////

			localProxy, err := New(
				// Local proxy URI.
				localProxyURI,

				// Upstream proxy URI.
				upstreamProxyURI,

				// PAC URI.
				pacURI,

				// PAC proxies credentials in standard URI format.
				tt.args.pacProxiesCredentials,

				// Logging settings.
				&Options{
					DNSURIs:        dnsURIs,
					LoggingOptions: loggingOptions,
				},
			)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("New() localProxy error = %v, wantErr %v", err, tt.wantErr)

					return
				}

				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("New() localProxy error = %v, expected %v", err, tt.wantErrType)

					return
				}

				return
			}

			// Both local `localProxy.LocalProxyURI` and `localProxy.UpstreamProxyURI`
			// changed, if `FORWARDER_LOCALPROXY_CREDENTIAL` or `FORWARDER_UPSTREAMPROXY_CREDENTIAL`
			// were set. This updates test vars.
			if localProxyURI != localProxy.LocalProxyURI {
				localProxyURI = localProxy.LocalProxyURI

				lPURI, err := url.ParseRequestURI(localProxyURI)
				if err != nil {
					t.Fatal("Failed to ParseRequestURI(localProxyURI)")
				}

				tt.args.localProxyURI = lPURI
			}

			if upstreamProxyURI != localProxy.UpstreamProxyURI {
				upstreamProxyURI = localProxy.UpstreamProxyURI

				lUURI, err := url.ParseRequestURI(upstreamProxyURI)
				if err != nil {
					t.Fatal("Failed to ParseRequestURI(upstreamProxyURI)")
				}

				tt.args.upstreamProxyURI = lUURI
			}

			go localProxy.Run()

			// Give enough time to start, and be ready.
			time.Sleep(1 * time.Second)

			//////
			// Upstream Proxy.
			//////

			if tt.preUpstreamFunc != nil {
				tt.preUpstreamFunc()
			}

			if upstreamProxyURI != "" {
				upstreamProxy, err := New(
					// Local proxy URI.
					upstreamProxyURI,

					// Upstream proxy URI.
					"",

					// PAC URI.
					"",

					// PAC proxies credentials in standard URI format.
					nil,

					// Logging settings.
					&Options{
						LoggingOptions: loggingOptions,
					},
				)
				if err != nil {
					if !tt.wantErr {
						t.Errorf("New() localProxy error = %v, wantErr %v", err, tt.wantErr)

						return
					}

					if !errors.Is(err, tt.wantErrType) {
						t.Errorf("New() localProxy error = %v, expected %v", err, tt.wantErrType)

						return
					}

					return
				}

				go upstreamProxy.Run()

				// Give enough time to start, and be ready.
				time.Sleep(1 * time.Second)
			}

			if tt.postUpstreamFunc != nil {
				tt.postUpstreamFunc()
			}

			//////
			// Client.
			//////

			l.Debuglnf("Client is using %s as proxy", tt.args.localProxyURI.Redacted())

			// Client's proxy settings.
			tr := &http.Transport{
				Proxy: http.ProxyURL(tt.args.localProxyURI),
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

			if tt.postFunc != nil {
				tt.postFunc()
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	//////
	// Target/end server.
	//////

	testServer := createMockedHTTPServer(http.StatusOK, "body", "")

	defer func() { testServer.Close() }()

	//////
	// Proxy.
	//////

	r, err := randomness.New(30000, 40000, 100, true)
	if err != nil {
		//nolint:gocritic
		log.Fatalln("Failed to create proxy.", err)
	}

	localProxyURI := URIBuilder(defaultProxyHostname, r.MustGenerate(), "", "")

	proxy, err := New(localProxyURI.String(), "", "", nil, nil)
	if err != nil {
		log.Fatalln("Failed to create proxy.", err)
	}

	go proxy.Run()

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
		_, _, _ = executeRequest(client, testServer.URL)
	}
}
