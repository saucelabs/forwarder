// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

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
// The implemented endpoints are:
// `/basic-auth/{user}/{passwd}`, `/delay/{seconds}`, `/status/{code}`, `/stream-bytes/{bytes}`.
func Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/basic-auth/", basicAuthHandler)
	m.HandleFunc("/delay/", delayHandler)
	m.HandleFunc("/status/", statusHandler)
	m.HandleFunc("/stream-bytes/", streamBytesHandler)
	m.HandleFunc("/events/", events)
	m.HandleFunc("/events.html", eventsHTML)
	m.HandleFunc("/ws/echo", wsEcho)
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

// delayHandler implements the /delay/{seconds} endpoint.
func delayHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/delay/"):]

	ms, ok := atoi(w, p)
	if !ok {
		return
	}

	t := time.NewTimer(time.Duration(ms) * time.Millisecond)
	select {
	case <-r.Context().Done():
		t.Stop()
	case <-t.C:
		w.WriteHeader(http.StatusOK)
	}
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

	io.CopyBuffer(w, &patternReader{
		Pattern: []byte("SauceLabs"),
		N:       int64(n),
	}, make([]byte, chunkSize))
}

func atoi(w http.ResponseWriter, s string) (int, bool) {
	v, err := strconv.Atoi(s)
	if err != nil {
		msg := fmt.Sprintf("invalid argument %q: %s", s, err)
		http.Error(w, msg, http.StatusBadRequest)
	}
	return v, err == nil
}
