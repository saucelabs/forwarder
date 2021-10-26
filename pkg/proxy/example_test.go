// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/saucelabs/forwarder/internal/logger"
	"github.com/saucelabs/randomness"
	"github.com/saucelabs/sypl/level"
)

// Complete, and complex example.
//
// client -> protected local proxy -> protected pac server - connection setup -> protected upstream proxy -> protected target.
func ExampleNew() {
	//////
	// Setup demo logger.
	//////

	// Only `stdout`, and `stderr`
	loggingOptions := &LoggingOptions{
		FileLevel: level.None.String(),
		FilePath:  "-",

		// Change to `Trace` for debugging, and demonstration purposes.
		Level: level.None.String(),
	}

	l := logger.Setup(loggingOptions)

	//////
	// Randomness automates port allocation, ensuring no collision happens
	// between tests, and examples.
	//////

	r, err := randomness.New(49000, 50000, 100, true)
	if err != nil {
		log.Fatalln("Failed to create randomness.", err)
	}

	//////
	// Target/end server.
	//////

	targetServer := createMockedHTTPServer(http.StatusOK, "body", "dXNlcjE6cGFzczE=")

	defer func() { targetServer.Close() }()

	targetServerURI, err := url.ParseRequestURI(targetServer.URL)
	if err != nil {
		//nolint:gocritic
		log.Fatalln("Failed to parse target server URL.", err)
	}

	targetServerURI.User = url.UserPassword("user1", "pass1")

	l.Debuglnf("Target/end server started @ %s", targetServerURI.Redacted())

	//////
	// PAC content.
	//////

	// Use `int(r.MustGenerate())` for testing purposes. Specify a port if using
	// a manual - external proxy (e.g.: NGINX). Good for debugging, and demo
	// purposes.
	upstreamProxyPort := int(r.MustGenerate())

	templateMap := map[string]int{
		"port": upstreamProxyPort,
	}

	var pacText strings.Builder
	_ = template.Must(template.New("pacTemplate").Parse(pacTemplate)).Execute(&pacText, templateMap)

	l.Debuglnf("PAC template parsed: \n%s", pacText.String())

	//////
	// PAC server.
	//////

	pacServer := createMockedHTTPServer(http.StatusOK, pacText.String(), "dXNlcjpwYXNz")

	defer func() { pacServer.Close() }()

	pacServerURI, err := url.ParseRequestURI(pacServer.URL)
	if err != nil {
		log.Fatalln("Failed to parse PAC server URL.", err)
	}

	pacServerURI.User = url.UserPassword("user", "pass")

	l.Debuglnf("PAC server started @ %s", pacServerURI.Redacted())

	//////
	// URL for both proxies, local, and upstream.
	//////

	// Local proxy.
	localProxyURI := URIBuilder(defaultProxyHostname, r.MustGenerate(), localProxyCredentialUsername, localProxyCredentialPassword)

	// Upstream proxy.
	upstreamProxyURI := URIBuilder(defaultProxyHostname, int64(upstreamProxyPort), upstreamProxyCredentialUsername, upstreamProxyCredentialPassword)

	//////
	// Local proxy.
	//
	// It's protected with Basic Auth. Upstream proxy will be automatically, and
	// dynamically setup via PAC, including credentials for proxies specified
	// in the PAC content.
	//////

	localProxy, err := New(
		// Local proxy URI.
		localProxyURI.String(),

		// Upstream proxy URI.
		"",

		// PAC URI.
		pacServerURI.String(),

		// PAC proxies credentials in standard URI format.
		[]string{upstreamProxyURI.String()},

		// Logging settings.
		&Options{
			LoggingOptions: loggingOptions,
		},
	)
	if err != nil {
		log.Fatalln("Failed to create proxy.", err)
	}

	go localProxy.Run()

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	//////
	// Upstream Proxy.
	//////

	upstreamProxy, err := New(
		// Local proxy URI.
		upstreamProxyURI.String(),

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
		log.Fatalln("Failed to create upstream proxy.", err)
	}

	go upstreamProxy.Run()

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	//////
	// Client.
	//////

	l.Debuglnf("Client is using %s as proxy", localProxyURI.Redacted())

	// Client's proxy settings.
	tr := &http.Transport{
		Proxy: http.ProxyURL(localProxyURI),
	}

	client := &http.Client{
		Transport: tr,
	}

	statusCode, body, err := executeRequest(client, targetServerURI.String())
	if err != nil {
		log.Fatalf("Failed to execute request: %v", err)
	}

	fmt.Println(statusCode)
	fmt.Println(body)

	// output:
	// 200
	// body
}

// Automatically retry port example.
func ExampleNew_automaticallyRetryPort() {
	if os.Getenv("FORWARDER_TEST_MODE") != "integration" {
		fmt.Println("true")

		return
	}

	//////
	// Randomness automates port allocation, ensuring no collision happens.
	//////

	r, err := randomness.New(55000, 65000, 100, true)
	if err != nil {
		log.Fatalln("Failed to create randomness.", err)
	}

	randomPort := r.MustGenerate()

	errored := false

	proxy1, err := New(fmt.Sprintf("http://0.0.0.0:%d", randomPort), "", "", nil, &Options{
		LoggingOptions: &LoggingOptions{
			Level:     "none",
			FileLevel: "none",
		},
	})
	if err != nil {
		errored = true
	}

	go proxy1.Run()

	time.Sleep(1 * time.Second)

	proxy2, err := New(fmt.Sprintf("http://0.0.0.0:%d", randomPort), "", "", nil, &Options{
		AutomaticallyRetryPort: true,

		LoggingOptions: &LoggingOptions{
			Level:     "none",
			FileLevel: "none",
		},
	})
	if err != nil {
		errored = true
	}

	go proxy2.Run()

	time.Sleep(1 * time.Second)

	fmt.Println(errored == false)

	// output:
	// true
}
