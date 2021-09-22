// Copyright 2021 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package proxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

// Demonstrates how to start a simple proxy. Flow:
// Client -> Proxy -> Target.
func ExampleNew() {
	//////
	// Target.
	//////

	// Mocked HTTP server.
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		if _, err := res.Write([]byte("body")); err != nil {
			log.Fatalln("Failed to write body.", err)
		}
	}))

	defer func() { testServer.Close() }()

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	//////
	// Proxy.
	//////

	proxyHost := "localhost:8080"

	proxy, err := New(proxyHost, "", "", "", nil)
	if err != nil {
		//nolint:gocritic
		log.Fatalln("Failed to create proxy.", err)
	}

	go proxy.Run()

	// Give enough time to start, and be ready.
	time.Sleep(1 * time.Second)

	//////
	// Client.
	//////

	proxyURL, err := url.Parse(fmt.Sprintf("http://%s", proxyHost))
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	// Client's proxy settings.
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: tr,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Failed to execute request: %v", err)
	}

	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Failed request, non-2xx code (%d): %s", response.StatusCode, data)
	}

	fmt.Println(response.StatusCode)
	fmt.Println(string(data))

	// output:
	// 200
	// body
}
