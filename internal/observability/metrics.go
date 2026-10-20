// Package observability provides metrics and logging helpers.
package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds application metrics.
type Metrics struct {
	RequestsTotal *prometheus.CounterVec
	RequestDur    *prometheus.HistogramVec
}

// NewMetrics registers and returns application metrics.
func NewMetrics() *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "flaggo_http_requests_total",
				Help: "Total HTTP requests by method, path, status.",
			},
			[]string{"method", "path", "status"},
		),
		RequestDur: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "flaggo_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}
	prometheus.MustRegister(m.RequestsTotal, m.RequestDur)
	return m
}

// Handler returns the Prometheus scrape handler.
func Handler() http.Handler {
	return promhttp.Handler()
}
