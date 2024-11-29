package mitmprom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/saucelabs/forwarder/internal/martian/mitm"
)

type CacheMetricsFunc func() mitm.CacheMetrics

type CacheMetricsCollector struct {
	inserts    *prometheus.Desc
	collisions *prometheus.Desc
	evictions  *prometheus.Desc
	removals   *prometheus.Desc
	hits       *prometheus.Desc
	misses     *prometheus.Desc
	metrics    CacheMetricsFunc
}

func NewCacheMetricsCollector(namespace string, f CacheMetricsFunc) *CacheMetricsCollector {
	return &CacheMetricsCollector{
		inserts: prometheus.NewDesc(
			namespace+"cache_inserts_total",
			"Number of cache inserts.",
			nil, nil,
		),
		collisions: prometheus.NewDesc(
			namespace+"cache_collisions_total",
			"Number of cache collisions.",
			nil, nil,
		),
		evictions: prometheus.NewDesc(
			namespace+"cache_evictions_total",
			"Number of cache evictions.",
			nil, nil,
		),
		removals: prometheus.NewDesc(
			namespace+"cache_removals_total",
			"Number of cache removals.",
			nil, nil,
		),
		hits: prometheus.NewDesc(
			namespace+"cache_hits_total",
			"Number of cache hits.",
			nil, nil,
		),
		misses: prometheus.NewDesc(
			namespace+"cache_misses_total",
			"Number of cache misses.",
			nil, nil,
		),
		metrics: f,
	}
}

func (c *CacheMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.inserts
	ch <- c.collisions
	ch <- c.evictions
	ch <- c.removals
	ch <- c.hits
	ch <- c.misses
}

func (c *CacheMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	m := c.metrics()
	ch <- prometheus.MustNewConstMetric(c.inserts, prometheus.CounterValue, float64(m.Inserts))
	ch <- prometheus.MustNewConstMetric(c.collisions, prometheus.CounterValue, float64(m.Collisions))
	ch <- prometheus.MustNewConstMetric(c.evictions, prometheus.CounterValue, float64(m.Evictions))
	ch <- prometheus.MustNewConstMetric(c.removals, prometheus.CounterValue, float64(m.Removals))
	ch <- prometheus.MustNewConstMetric(c.hits, prometheus.CounterValue, float64(m.Hits))
	ch <- prometheus.MustNewConstMetric(c.misses, prometheus.CounterValue, float64(m.Misses))
}
