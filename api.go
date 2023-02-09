// Copyright 2022 The forwarder Authors. All rights reserved.
// Use of this source code is governed by a MPL
// license that can be found in the LICENSE file.

package forwarder

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/saucelabs/forwarder/internal/version"
)

type server interface {
	Addr() string
}

// APIHandler serves API endpoints.
// It provides health and readiness endpoints prometheus metrics, and pprof debug endpoints.
type APIHandler struct {
	mux    *http.ServeMux
	server server
	config string
	script string
}

func NewAPIHandler(r prometheus.Gatherer, s server, config, pac string) *APIHandler {
	m := http.NewServeMux()
	a := &APIHandler{
		mux:    m,
		server: s,
		config: config,
		script: pac,
	}
	m.HandleFunc("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}).ServeHTTP)
	m.HandleFunc("/healthz", a.healthz)
	m.HandleFunc("/readyz", a.readyz)
	m.HandleFunc("/configz", a.configz)
	m.HandleFunc("/pac", a.pac)
	m.HandleFunc("/version", a.version)

	m.HandleFunc("/debug/pprof/", pprof.Index)
	m.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return a
}

func (h *APIHandler) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("OK"))
}

func (h *APIHandler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.server.Addr() != "" {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Service Unavailable"))
	}
}

func (h *APIHandler) configz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(h.config))
}

func (h *APIHandler) pac(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	w.Write([]byte(h.script))
}

func (h *APIHandler) version(w http.ResponseWriter, r *http.Request) {
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

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
