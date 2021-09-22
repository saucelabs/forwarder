// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/saucelabs/forwarder/internal/credential"
	"github.com/saucelabs/forwarder/internal/randomness"
)

const (
	cred                  = "u123:p123"
	host                  = "localhost:8080"
	parentProxyCredential = "u456:p456"
	parentProxyURL        = "http://localhost:8080"
)

var loggingOptions = &LoggingOptions{
	Level: "info", // NOTE: Set it to `trace` to debug problems.,
}

func TestNew(t *testing.T) {
	type args struct {
		host                  string
		cred                  string
		parentProxyURL        string
		parentProxyCredential string
		logLevel              *LoggingOptions
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrType error
	}{
		{
			name: "Should work - setting host",
			args: args{
				host:     host,
				logLevel: loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should work - setting host, and basic auth",
			args: args{
				host:     host,
				cred:     cred,
				logLevel: loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should work - setting host, basic auth, and parent proxy",
			args: args{
				host:           host,
				cred:           cred,
				parentProxyURL: parentProxyURL,
				logLevel:       loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should work - setting host, basic auth, parent proxy, and parent basic auth",
			args: args{
				host:                  host,
				cred:                  cred,
				parentProxyURL:        parentProxyURL,
				parentProxyCredential: parentProxyCredential,
				logLevel:              loggingOptions,
			},
			wantErr: false,
		},
		{
			name: "Should fail - missing host",
			args: args{
				logLevel: loggingOptions,
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyHost,
		},
		{
			name: "Should fail - invalid host",
			args: args{
				host:     ":",
				logLevel: loggingOptions,
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyHost,
		},
		{
			name: "Should fail - invalid cred",
			args: args{
				host:     host,
				cred:     "::",
				logLevel: loggingOptions,
			},
			wantErr:     true,
			wantErrType: credential.ErrUsernamePasswordRequired,
		},
		{
			name: "Should fail - invalid parent proxy URL",
			args: args{
				host:           host,
				cred:           cred,
				parentProxyURL: "localhost",
				logLevel:       loggingOptions,
			},
			wantErr:     true,
			wantErrType: ErrInvalidProxyURL,
		},
		{
			name: "Should fail - invalid parent proxy cred",
			args: args{
				host:                  host,
				cred:                  cred,
				parentProxyURL:        parentProxyURL,
				parentProxyCredential: "::",
				logLevel:              loggingOptions,
			},
			wantErr:     true,
			wantErrType: credential.ErrUsernamePasswordRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.host, tt.args.cred, tt.args.parentProxyURL, tt.args.parentProxyCredential, tt.args.logLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err != nil) == tt.wantErr {
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("New() error is = %v, wantErrType %v", err, tt.wantErrType)
				}
			}
		})
	}
}

func TestNew_Run(t *testing.T) {
	// Mocked HTTP server.
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		if _, err := res.Write([]byte("body")); err != nil {
			t.Fatal("Failed to write body", err)
		}
	}))

	defer func() { testServer.Close() }()

	time.Sleep(1 * time.Second)

	testURL := testServer.URL

	// Live test. Test calls to non-localhost. It matters because non-localhost
	// uses the proxy settings in the Transport, while localhost calls, bypass
	// it.
	if os.Getenv("PROXY_TEST_MODE") == "integration" {
		testURL = "https://httpbin.org/status/200"
	}

	type args struct {
		min, max              int
		minP, maxP            int
		parentProxy           bool
		cred                  string
		parentProxyCredential string
		logLevel              *LoggingOptions
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrType error
		testURL     string
	}{
		{
			name: "Should work - basic proxy",
			args: args{
				min:      12000,
				max:      13000,
				logLevel: loggingOptions,
			},
			wantErr: false,
			testURL: testURL,
		},
		{
			name: "Should work - setting host, and basic auth",
			args: args{
				min:      14000,
				max:      15000,
				cred:     cred,
				logLevel: loggingOptions,
			},
			wantErr: false,
			testURL: testURL,
		},
		{
			name: "Should work - setting host, basic auth, and parent proxy",
			args: args{
				min:         16000,
				max:         17000,
				cred:        cred,
				minP:        18000,
				maxP:        19000,
				parentProxy: true,
				logLevel:    loggingOptions,
			},
			wantErr: false,
			testURL: testURL,
		},
		{
			name: "Should work - setting host, basic auth, parent proxy, and parent basic auth",
			args: args{
				min:                   20000,
				max:                   21000,
				cred:                  cred,
				minP:                  22000,
				maxP:                  23000,
				parentProxy:           true,
				parentProxyCredential: parentProxyCredential,
				logLevel:              loggingOptions,
			},
			wantErr: false,
			testURL: testURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//////
			// Random ports for both proxy, and parent proxy - if needed.
			//////

			parentPort, err := randomness.New(tt.args.minP, tt.args.maxP, 10, true)
			if err != nil {
				t.Fatal(err)
			}

			parentProxyHost := ""
			parentProxyURL := ""

			if tt.args.parentProxy {
				parentProxyHost = fmt.Sprintf("localhost:%d", parentPort.MustGenerate())
				parentProxyURL = fmt.Sprintf("http://%s", parentProxyHost)
			}

			port, err := randomness.New(tt.args.min, tt.args.max, 10, true)
			if err != nil {
				t.Fatal(err)
			}

			host := fmt.Sprintf("localhost:%d", port.MustGenerate())

			//////
			// Proxy.
			//////

			proxy, err := New(
				host,
				tt.args.cred,
				parentProxyURL,
				tt.args.parentProxyCredential,
				tt.args.logLevel,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (err != nil) == tt.wantErr {
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("New() error is = %v, wantErrType %v", err, tt.wantErrType)
				}
			}

			go proxy.Run()

			time.Sleep(1 * time.Second)

			//////
			// URL to request.
			//////

			u := &url.URL{
				Scheme: "http",
				Host:   host,
			}

			// Set basic auth - if needed.
			if tt.args.cred != "" {
				u.User = url.UserPassword(proxy.Credential.Username, proxy.Credential.Password)
			}

			//////
			// Parent proxy.
			//////

			if tt.args.parentProxy {
				parentProxy, err := New(
					parentProxyHost,
					tt.args.parentProxyCredential,
					"",
					"",
					tt.args.logLevel,
				)

				if (err != nil) != tt.wantErr {
					t.Errorf("New() parent proxy error = %v, wantErr %v", err, tt.wantErr)

					return
				}

				go parentProxy.Run()

				time.Sleep(1 * time.Second)
			}

			//////
			// Execute request.
			//////

			tr := &http.Transport{
				Proxy: http.ProxyURL(u),
			}

			client := &http.Client{
				Transport: tr,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			request, err := http.NewRequestWithContext(ctx, http.MethodGet, tt.testURL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			response, err := client.Do(request)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			defer response.Body.Close()

			data, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}

			if response.StatusCode != http.StatusOK {
				t.Fatalf("Failed request, non-2xx code (%d): %s", response.StatusCode, data)
			}
		})
	}
}
