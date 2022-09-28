// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forwarder

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
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
	targetServer := httpServerStub("body", "dXNlcjE6cGFzczE=", namedStdLogger("target"))

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
	pacServer := httpServerStub(pacText.String(), "dXNlcjpwYXNz", namedStdLogger("pac"))

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
	localProxyURI := newProxyURL(r.MustGenerate(), localProxyCredentialUsername, localProxyCredentialPassword)

	// Upstream proxy.
	upstreamProxyURI := newProxyURL(int64(upstreamProxyPort), upstreamProxyCredentialUsername, upstreamProxyCredentialPassword)

	//////
	// Local proxy.
	//
	// It's protected with Basic Auth. Upstream proxy URL and credentials are determined
	// per URL via PAC.
	//////

	c := ProxyConfig{
		LocalProxyURI:         localProxyURI,
		PACURI:                pacServerURI,
		PACProxiesCredentials: []string{upstreamProxyURI.String()},
		ProxyLocalhost:        true,
	}
	localProxy, err := NewProxy(&c, log)
	if err != nil {
		panic(err)
	}

	go localProxy.MustRun()

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	//////
	// Upstream Proxy.
	//////

	upstreamProxy, err := NewProxy(&ProxyConfig{
		LocalProxyURI:  upstreamProxyURI,
		ProxyLocalhost: true,
	}, log)
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

	body, err := assertRequest(client, targetServerURI.String(), http.StatusOK)
	if err != nil {
		panic(err)
	}

	fmt.Println(body)

	// output:
	// body
}
