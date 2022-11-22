package httpbin

import (
	"fmt"
	"io"
	"math/rand"
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
	return m
}

// basicAuthHandler implements the /basic-auth/{user}/{passwd} endpoint.
// See https://httpbin.org/#/Auth/get_basic_auth__user___passwd_
func basicAuthHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/basic-auth/"):]

	user, pass, ok := strings.Cut(p, "/")
	if !ok {
		msg := fmt.Sprintf("Invalid path %q, expected ≤user>/<password>", p)
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

	s, ok := atoi(w, p)
	if !ok {
		return
	}
	if s > 10 {
		s = 10
	}

	time.Sleep(time.Duration(s) * time.Second)
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
}

var rnd = rand.NewSource(time.Now().Unix())

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
	io.CopyBuffer(w, io.LimitReader(rand.New(rnd), int64(n)), make([]byte, chunkSize))
}

func atoi(w http.ResponseWriter, s string) (int, bool) {
	v, err := strconv.Atoi(s)
	if err != nil {
		msg := fmt.Sprintf("Invalid argument %q: %s", s, err)
		http.Error(w, msg, http.StatusBadRequest)
	}
	return v, err == nil
}
