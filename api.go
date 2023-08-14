// Copyright 2023 Sauce Labs Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sort"
	"text/template"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/saucelabs/forwarder/internal/version"
)

// APIHandler serves API endpoints.
// It provides health and readiness endpoints prometheus metrics, and pprof debug endpoints.
type APIHandler struct {
	mux    *http.ServeMux
	ready  func(ctx context.Context) bool
	config string
	script string

	patterns []string
}

func NewAPIHandler(r prometheus.Gatherer, ready func(ctx context.Context) bool, config, pac string) *APIHandler {
	m := http.NewServeMux()
	a := &APIHandler{
		mux:    m,
		ready:  ready,
		config: config,
		script: pac,
	}

	var indexPatterns []string
	handleFunc := func(pattern string, handler func(http.ResponseWriter, *http.Request)) {
		indexPatterns = append(indexPatterns, pattern)
		m.HandleFunc(pattern, handler)
	}

	handleFunc("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}).ServeHTTP)
	handleFunc("/healthz", a.healthz)
	handleFunc("/readyz", a.readyz)
	handleFunc("/configz", a.configz)
	handleFunc("/pac", a.pac)
	handleFunc("/version", a.version)

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
	if h.ready(r.Context()) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Service Unavailable"))
	}
}

func (h *APIHandler) configz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(h.config))
}

func (h *APIHandler) pac(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	w.Write([]byte(h.script))
}

func (h *APIHandler) version(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	v := struct {
		Version string `json:"version"`
		Time    string `json:"time"`
		Commit  string `json:"commit"`

		GoArch    string `json:"go_arch"`
		GOOS      string `json:"go_os"`
		GoVersion string `json:"go_version"`
	}{
		Version: version.Version,
		Time:    version.Time,
		Commit:  version.Commit,

		GoArch:    runtime.GOARCH,
		GOOS:      runtime.GOOS,
		GoVersion: runtime.Version(),
	}
	json.NewEncoder(w).Encode(v) //nolint // ignore error
}

const indexTemplate = `<!DOCTYPE html>
<html>
<head>
<title>Forwarder {{.Version}}</title>
</head>
<body>
<h1>Forwarder {{.Version}}</h1>
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
		Version  string
		Patterns []string
	}{
		Version:  version.Version,
		Patterns: h.patterns,
	}); err != nil {
		w.Write([]byte(err.Error()))
	}

	buf.WriteTo(w) //nolint:errcheck // ignore error
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
