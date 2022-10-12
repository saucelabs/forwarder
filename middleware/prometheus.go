package middleware

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	_          = iota // ignore first value by assigning to blank identifier
	kb float64 = 1 << (10 * iota)
	mb
)

var sizeBuckets = []float64{ //nolint:gochecknoglobals // this is a global variable by design
	1 * kb,
	2 * kb,
	5 * kb,
	10 * kb,
	100 * kb,
	500 * kb,
	1 * mb,
	2.5 * mb,
	5 * mb,
	10 * mb,
}

// Prometheus is a middleware that collects metrics about the HTTP requests and responses.
// Unlike the promhttp.InstrumentHandler* chaining, this middleware creates only one delegator per request.
// It partitions the metrics by HTTP status code, HTTP method, destination host name and source IP.
type Prometheus struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
}

func NewPrometheus(r prometheus.Registerer, namespace string) *Prometheus {
	f := promauto.With(r)
	l := []string{"code", "method", "host", "source"}

	return &Prometheus{
		requestsTotal: f.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests processed.",
		}, l),
		requestDuration: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, l),
		requestSize: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
			Buckets:   sizeBuckets,
		}, l),
		responseSize: f.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
			Buckets:   sizeBuckets,
		}, l),
	}
}

func (p *Prometheus) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqSize := computeApproximateRequestSize(r)

		start := time.Now()
		d := newDelegator(w, nil)
		h.ServeHTTP(d, r)
		elapsed := float64(time.Since(start)) / float64(time.Second)

		src, _, _ := net.SplitHostPort(r.RemoteAddr) //nolint:errcheck // ignore error
		lv := [4]string{strconv.Itoa(d.Status()), r.Method, r.Host, src}

		p.requestsTotal.WithLabelValues(lv[:]...).Inc()
		p.requestDuration.WithLabelValues(lv[:]...).Observe(elapsed)
		p.requestSize.WithLabelValues(lv[:]...).Observe(float64(reqSize))
		p.responseSize.WithLabelValues(lv[:]...).Observe(float64(d.Written()))
	})
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
