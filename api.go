package forwarder

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		mux:   m,
		proxy: proxy,
	}
	m.HandleFunc("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}).ServeHTTP)
	m.HandleFunc("/healthz", a.healthz)
	m.HandleFunc("/readyz", a.readyz)
	m.HandleFunc("/configz", a.configz)
	m.HandleFunc("/pac", a.pac)
	m.HandleFunc("/debug/pprof/", pprof.Index)
	m.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return a
}

func (h *APIHandler) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK")) //nolint:errcheck // ignore error
}

func (h *APIHandler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.proxy.Addr() != "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK")) //nolint:errcheck // ignore error
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
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

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
