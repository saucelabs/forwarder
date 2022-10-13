package forwarder

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// APIHandler returns http.Handler serving API endpoints.
// It provides health and readiness endpoints prometheus metrics, and pprof debug endpoints.
func APIHandler(s *HTTPServer, r prometheus.Gatherer) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/v1/health", healthHandler)
	m.HandleFunc("/api/v1/ready", readyHandler(s))

	m.HandleFunc("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}).ServeHTTP)

	m.HandleFunc("/debug/pprof/", pprof.Index)
	m.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	m.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return m
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK")) //nolint:errcheck // ignore error
}

func readyHandler(s *HTTPServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Addr() != "" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK")) //nolint:errcheck // ignore error
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service Unavailable")) //nolint:errcheck // ignore error
		}
	}
}
