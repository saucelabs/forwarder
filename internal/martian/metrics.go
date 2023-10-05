package martian

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	inFlight prometheus.Gauge
}

func newMetrics(r prometheus.Registerer, namespace string) *metrics {
	if r == nil {
		r = prometheus.NewRegistry() // This registry will be discarded.
	}
	f := promauto.With(r)

	return &metrics{
		inFlight: f.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "in_flight_requests",
			Help:      "Current number of requests being served.",
		}),
	}
}

func (m *metrics) RequestReceived() {
	m.inFlight.Inc()
}

func (m *metrics) RequestDone() {
	m.inFlight.Dec()
}
