// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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

func defaultLoggingOptions() *LoggingOptions {
	return &LoggingOptions{
		Level: defaultProxyLoggingLevel, // NOTE: Set it to `trace` to debug problems.
	}
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

// Creates a mocked HTTP server. Don't forget to defer close it!
//
//nolint:unparam //`statusCode` always receives `http.StatusOK` (`200`)
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

func executeRequest(client *http.Client, uri string) (statusCode int, body string, err error) {
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

	data, err := io.ReadAll(response.Body)
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

func TestParseSiteCredentials(t *testing.T) {
	tests := []struct {
		name     string
		in       []string
		hostPort map[string]string
		port     map[string]string
		host     map[string]string
		global   string
		err      bool
	}{
		{
			name: "Valid with schema",
			in:   []string{"https://user:pass@abc"},
			hostPort: map[string]string{
				"abc": "dXNlcjpwYXNz",
			},
			host: map[string]string{},
			port: map[string]string{},
		},
		{
			name: "Empty user",
			in:   []string{":pass@abc"},
			err:  true,
		},
		{
			name: "Empty password",
			in:   []string{"user:@abc"},
			err:  true,
		},
		{
			name: "Missing password",
			in:   []string{"user@abc"},
			err:  true,
		},
		{
			name: "Missing host",
			in:   []string{"user:pass"},
			err:  true,
		},
		{
			name: "Valid host",
			in:   []string{"user:pass@abc"},
			hostPort: map[string]string{
				"abc": "dXNlcjpwYXNz",
			},
			host: map[string]string{},
			port: map[string]string{},
		},
		{
			name: "Valid host+port",
			in:   []string{"user:pass@abc:123"},
			hostPort: map[string]string{
				"abc:123": "dXNlcjpwYXNz",
			},
			host: map[string]string{},
			port: map[string]string{},
		},
		{
			name: "Wildcard host",
			in:   []string{"user:pass@*:123"},
			port: map[string]string{
				"123": "dXNlcjpwYXNz",
			},
			host:     map[string]string{},
			hostPort: map[string]string{},
		},
		{
			name: "Wildcard port",
			in:   []string{"user:pass@abc:0"},
			host: map[string]string{
				"abc": "dXNlcjpwYXNz",
			},
			hostPort: map[string]string{},
			port:     map[string]string{},
		},
		{
			name:     "Global wildcard",
			in:       []string{"user:pass@*:0"},
			global:   "dXNlcjpwYXNz",
			hostPort: map[string]string{},
			host:     map[string]string{},
			port:     map[string]string{},
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			hostPort, host, port, global, err := parseSiteCredentials(tc.in)

			if (err == nil) == tc.err {
				t.Fatalf("Unexpected error condition: %s", err)
			}

			diff := cmp.Diff(tc.hostPort, hostPort)
			if diff != "" {
				t.Fatalf(diff)
			}

			diff = cmp.Diff(tc.host, host)
			if diff != "" {
				t.Fatalf(diff)
			}
			diff = cmp.Diff(tc.port, port)
			if diff != "" {
				t.Fatalf(diff)
			}
			diff = cmp.Diff(tc.global, global)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}

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
		siteCredentials       []string
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
				loggingOptions: defaultLoggingOptions(),
			},
			wantErr: false,
		},
		{
			name: "Should work - local proxy with site auth",
			args: args{
				localProxyURI: URIBuilder(
					defaultProxyHostname,
					r.MustGenerate(),
					"",
					"",
				),
				loggingOptions:  defaultLoggingOptions(),
				siteCredentials: []string{},
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
				loggingOptions: defaultLoggingOptions(),
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
				loggingOptions: defaultLoggingOptions(),
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
				loggingOptions: defaultLoggingOptions(),
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
				loggingOptions: defaultLoggingOptions(),
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
				loggingOptions: defaultLoggingOptions(),
			},
			wantErr:     true,
			wantErrType: ErrInvalidOrParentOrPac,
		},
		{
			name: "Should fail - missing local proxy URI",
			args: args{
				loggingOptions: defaultLoggingOptions(),
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyParams,
		},
		{
			name: "Should fail - invalid local proxy URI",
			args: args{
				localProxyURI:  URIBuilder("", 0, "", ""),
				loggingOptions: defaultLoggingOptions(),
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyParams,
		},
		{
			name: "Should fail - invalid upstream proxy URI",
			args: args{
				localProxyURI:    URIBuilder(defaultProxyHostname, r.MustGenerate(), "", ""),
				upstreamProxyURI: URIBuilder("", 0, "", ""),
				loggingOptions:   defaultLoggingOptions(),
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyParams,
		},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			if tc.preFunc != nil {
				tc.preFunc()
			}

			// Re-set the value allowing per-test case setting.
			l := logger.Setup(tc.args.loggingOptions)

			//////
			// Target/end server.
			//////

			targetCreds := ""
			if tc.args.siteCredentials != nil {
				targetCreds = base64.
					StdEncoding.
					EncodeToString([]byte("user:pass"))
			}
			targetServer := createMockedHTTPServer(http.StatusOK, "body", targetCreds)

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

			l.Debuglnf("Target/end server @ %s", targetServerURL)

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

			localProxy, err := New(
				// Local proxy URI.
				localProxyURI,

				// Upstream proxy URI.
				upstreamProxyURI,

				// PAC URI.
				pacURI,

				// PAC proxies credentials in standard URI format.
				tc.args.pacProxiesCredentials,

				// Logging settings.
				&Options{
					DNSURIs:        dnsURIs,
					LoggingOptions: defaultLoggingOptions(),
					// site credentials in standard URI format.
					SiteCredentials: siteCredentials,
				},
			)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("New() localProxy error = %v, wantErr %v", err, tc.wantErr)

					return
				}

				if !errors.Is(err, tc.wantErrType) {
					t.Errorf("New() localProxy error = %v, expected %v", err, tc.wantErrType)

					return
				}

				return
			}

			// Both local `localProxy.LocalProxyURI` and `localProxy.UpstreamProxyURI`
			// changed, if `FORWARDER_LOCALPROXY_AUTH` or `FORWARDER_UPSTREAMPROXY_AUTH`
			// were set. This updates test vars.
			if localProxyURI != localProxy.LocalProxyURI {
				localProxyURI = localProxy.LocalProxyURI

				lPURI, err := url.ParseRequestURI(localProxyURI)
				if err != nil {
					t.Fatal("Failed to ParseRequestURI(localProxyURI)")
				}

				tc.args.localProxyURI = lPURI
			}

			if upstreamProxyURI != localProxy.UpstreamProxyURI {
				upstreamProxyURI = localProxy.UpstreamProxyURI

				lUURI, err := url.ParseRequestURI(upstreamProxyURI)
				if err != nil {
					t.Fatal("Failed to ParseRequestURI(upstreamProxyURI)")
				}

				tc.args.upstreamProxyURI = lUURI
			}

			go localProxy.Run()

			// Give enough time to start, and be ready.
			time.Sleep(1 * time.Second)

			//////
			// Upstream Proxy.
			//////

			if tc.preUpstreamFunc != nil {
				tc.preUpstreamFunc()
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
						LoggingOptions: defaultLoggingOptions(),
					},
				)
				if err != nil {
					if !tc.wantErr {
						t.Errorf("New() localProxy error = %v, wantErr %v", err, tc.wantErr)

						return
					}

					if !errors.Is(err, tc.wantErrType) {
						t.Errorf("New() localProxy error = %v, expected %v", err, tc.wantErrType)

						return
					}

					return
				}

				go upstreamProxy.Run()

				// Give enough time to start, and be ready.
				time.Sleep(1 * time.Second)
			}

			if tc.postUpstreamFunc != nil {
				tc.postUpstreamFunc()
			}

			//////
			// Client.
			//////

			l.Debuglnf("Client is using %s as proxy", tc.args.localProxyURI.Redacted())

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

	proxy, err := New(localProxyURI.String(), "", "", nil,
		&Options{
			LoggingOptions: defaultLoggingOptions(),
		})
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
		_, _, err := executeRequest(client, testServer.URL)
		if err != nil {
			b.Fatal(err)
		}
	}
}
