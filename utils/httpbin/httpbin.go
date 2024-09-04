// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package httpbin

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/saucelabs/forwarder/middleware"
)

// Handler returns http.Handler that implements elements of httpbin.org API.
func Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/basic-auth/", basicAuthHandler)
	m.HandleFunc("/delay/", delayHandler)
	m.HandleFunc("/status/", statusHandler)
	m.HandleFunc("/stream-bytes/", streamBytesHandler)
	m.HandleFunc("/count-bytes/", countBytesHandler)
	m.HandleFunc("/events/", events)
	m.HandleFunc("/events.html", eventsHTML)
	m.HandleFunc("/ws/echo", wsEcho)
	m.HandleFunc("/ws.html", wsHTML)
	m.HandleFunc("/headers/", headersHandler)
	return m
}

// basicAuthHandler implements the /basic-auth/{user}/{passwd} endpoint.
// See https://httpbin.org/#/Auth/get_basic_auth__user___passwd_
func basicAuthHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/basic-auth/"):]

	user, pass, ok := strings.Cut(p, "/")
	if !ok {
		msg := fmt.Sprintf("invalid path %q, expected ≤user>/<password>", p)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	mw := middleware.NewBasicAuth()
	if !mw.AuthenticatedRequest(r, user, pass) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// delayHandler implements the /delay/{milliseconds} endpoint.
func delayHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/delay/"):]

	ms, ok := atoi(w, p)
	if !ok {
		return
	}

	t := time.NewTimer(time.Duration(ms) * time.Millisecond)
	defer t.Stop()

	select {
	case <-r.Context().Done():
		t.Stop()
	case <-t.C:
	}

	w.WriteHeader(http.StatusOK)
}

// statusHandler implements the /status/{code} endpoint.
// See https://httpbin.org/#/Status_codes
func statusHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/status/"):]

	c, ok := atoi(w, p)
	if !ok {
		return
	}
	w.WriteHeader(c)

	q := r.URL.Query()
	if b := q.Get("body"); b == "true" {
		w.Write([]byte(http.StatusText(c)))
	}
}

// streamBytesHandler implements the /stream-bytes/{bytes} endpoint.
// See https://httpbin.org/#/Dynamic_data/get_stream_bytes__n_
func streamBytesHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/stream-bytes/"):]

	n, ok := atoi(w, p)
	if !ok {
		return
	}

	q := r.URL.Query()
	chunkSize := 10 * 1024
	if cs := q.Get("chunk_size"); cs != "" {
		chunkSize, ok = atoi(w, cs)
		if !ok {
			return
		}
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	io.CopyBuffer(w, &patternReader{ //nolint:errcheck // best effort
		Pattern: []byte("SauceLabs"),
		N:       int64(n),
	}, make([]byte, chunkSize))
}

// countBytesHandler implements the /count-bytes/ endpoint.
// It reads the request body and sends back the number of bytes read in a `Body-Size` header.
func countBytesHandler(w http.ResponseWriter, r *http.Request) {
	n, _ := io.Copy(io.Discard, r.Body) //nolint:errcheck // best effort
	w.Header().Set("Body-Size", strconv.FormatInt(n, 10))
	w.WriteHeader(http.StatusOK)
}

func atoi(w http.ResponseWriter, s string) (int, bool) {
	v, err := strconv.Atoi(s)
	if err != nil {
		msg := fmt.Sprintf("invalid argument %q: %s", s, err)
		http.Error(w, msg, http.StatusBadRequest)
		return 0, false
	}
	return v, true
}

func headersHandler(w http.ResponseWriter, r *http.Request) {
	for k, vv := range r.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
}
