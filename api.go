// Copyright 2022-2024 Sauce Labs Inc., all rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"bytes"
	"context"
	"net/http"
	"net/http/pprof"
	"sort"
	"text/template"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// APIUnixSocket is the path to the Unix socket for the API server.
// It is currently only used in containerized environments.
const APIUnixSocket = "/tmp/forwarder.sock"

// APIHandler serves API endpoints.
// It provides health and readiness endpoints prometheus metrics, and pprof debug endpoints.
type APIHandler struct {
	mux   *http.ServeMux
	ready func(ctx context.Context) bool

	title    string
	patterns []string
}

type APIEndpoint struct {
	Path    string
	Handler http.Handler
}

func NewAPIHandler(title string, r prometheus.Gatherer, ready func(ctx context.Context) bool, extraEndpoints ...APIEndpoint) *APIHandler {
	m := http.NewServeMux()
	a := &APIHandler{
		mux:   m,
		ready: ready,
		title: title,
	}

	var indexPatterns []string
	handleFunc := func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		indexPatterns = append(indexPatterns, pattern)
		m.HandleFunc(pattern, handler)
	}

	handleFunc("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{
		DisableCompression: true,
		EnableOpenMetrics:  true,
	}).ServeHTTP)
	handleFunc("/healthz", a.healthz)
	handleFunc("/readyz", a.readyz)

	for _, e := range extraEndpoints {
		handleFunc(e.Path, e.Handler.ServeHTTP)
	}

	handleFunc("/debug/pprof/", pprof.Index)
	m.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.HandleFunc("/debug/pprof/trace", pprof.Trace)

	sort.Strings(indexPatterns)
	a.patterns = indexPatterns
	m.HandleFunc("/", a.index)

	return a
}

func (h *APIHandler) healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func (h *APIHandler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.ready == nil || h.ready(r.Context()) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Service Unavailable"))
	}
}

const indexTemplate = `<!DOCTYPE html>
<html>
<head>
<title>{{.Title}}</title>
</head>
<body>
<h1>{{.Title}}</h1>
<ul>
{{range .Patterns}}
<li><a href="{{.}}">{{.}}</a></li>
{{end}}
</ul>
</body>
</html>
`

func (h *APIHandler) index(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")

	t, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, struct {
		Title    string
		Patterns []string
	}{
		Title:    h.title,
		Patterns: h.patterns,
	}); err != nil {
		w.Write([]byte(err.Error()))
	}

	buf.WriteTo(w) //nolint:errcheck // ignore error
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
