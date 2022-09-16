// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/saucelabs/randomness"
)

// Complete, and complex example.
//
// client -> protected local proxy -> protected pac server - connection setup -> protected upstream proxy -> protected target.
func ExampleNew() {
	// Logger
	log := nopLogger{}

	//////
	// Randomness automates port allocation, ensuring no collision happens
	// between tests, and examples.
	//////

	r, err := randomness.New(49000, 50000, 100, true)
	if err != nil {
		panic(err)
	}

	//////
	// Target/end server.
	//////

	// Create a protected HTTP server. user1:pass1 base64-encoded is dXNlcjE6cGFzczE=.
	targetServer := createMockedHTTPServer(http.StatusOK, "body", "dXNlcjE6cGFzczE=", namedStdLogger("target"))

	defer func() { targetServer.Close() }()

	targetServerURI, err := url.ParseRequestURI(targetServer.URL)
	if err != nil {
		panic(err)
	}

	targetServerURI.User = url.UserPassword("user1", "pass1")

	log.Debugf("Target/end server started @ %s", targetServerURI.Redacted())

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
	if err := template.Must(template.New("pacTemplate").Parse(pacTemplate)).Execute(&pacText, templateMap); err != nil {
		panic(err)
	}

	log.Debugf("PAC template parsed: \n%s", pacText.String())

	//////
	// PAC server.
	//////

	// Start a protected server (user:pass) serving PAC file.
	pacServer := createMockedHTTPServer(http.StatusOK, pacText.String(), "dXNlcjpwYXNz", namedStdLogger("pac"))

	defer func() { pacServer.Close() }()

	pacServerURI, err := url.ParseRequestURI(pacServer.URL)
	if err != nil {
		panic(err)
	}

	pacServerURI.User = url.UserPassword("user", "pass")

	log.Debugf("PAC server started @ %s", pacServerURI.Redacted())

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
	// It's protected with Basic Auth. Upstream proxy URL and credentials are determined
	// per URL via PAC.
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
		&Options{},
		log,
	)
	if err != nil {
		panic(err)
	}

	go localProxy.MustRun()

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
		&Options{},
		log,
	)
	if err != nil {
		panic(err)
	}

	go upstreamProxy.MustRun()

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	//////
	// Client.
	//////

	log.Debugf("Client is using %s as proxy", localProxyURI.Redacted())

	// Client's proxy settings.
	tr := &http.Transport{
		Proxy: http.ProxyURL(localProxyURI),
	}

	client := &http.Client{
		Transport: tr,
	}

	statusCode, body, err := executeRequest(client, targetServerURI.String())
	if err != nil {
		panic(err)
	}

	fmt.Println(statusCode)
	fmt.Println(body)

	// output:
	// 200
	// body
}

// Automatically retry port example.
func ExampleNew_automaticallyRetryPort() {
	// Logger
	log := nopLogger{}

	if os.Getenv("FORWARDER_TEST_MODE") != "integration" {
		fmt.Println("true")

		return
	}

	//////
	// Randomness automates port allocation, ensuring no collision happens.
	//////

	r, err := randomness.New(55000, 65000, 100, true)
	if err != nil {
		panic(err)
	}

	randomPort := r.MustGenerate()

	errored := false

	proxy1, err := New(fmt.Sprintf("http://0.0.0.0:%d", randomPort), "", "", nil, &Options{}, log)
	if err != nil {
		errored = true
	}

	go proxy1.MustRun()

	time.Sleep(1 * time.Second)

	proxy2, err := New(fmt.Sprintf("http://0.0.0.0:%d", randomPort), "", "", nil, &Options{AutomaticallyRetryPort: true}, log)
	if err != nil {
		errored = true
	}

	go proxy2.MustRun()

	time.Sleep(1 * time.Second)

	fmt.Println(errored == false)

	// output:
	// true
}
