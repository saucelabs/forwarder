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

// APIHandler serves API endpoints.
// It provides health and readiness endpoints prometheus metrics, and pprof debug endpoints.
type APIHandler struct {
	mux    *http.ServeMux
	proxy  *HTTPServer
	script string
}

func NewAPIHandler(r prometheus.Gatherer, proxy *HTTPServer, pac string) *APIHandler {
	m := http.NewServeMux()
	a := &APIHandler{
		mux:    m,
		proxy: proxy,
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
	w.Write([]byte("OK")) //nolint:errcheck // ignore error
}

func (h *APIHandler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.proxy.Addr() != "" {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK")) //nolint:errcheck // ignore error
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Service Unavailable")) //nolint:errcheck // ignore error
	}
}

func (h *APIHandler) configz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *APIHandler) pac(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	w.Write([]byte(h.script)) //nolint:errcheck // ignore it
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
	json.NewEncoder(w).Encode(v) //nolint // ignore it
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
