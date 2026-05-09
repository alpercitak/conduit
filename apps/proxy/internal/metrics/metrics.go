package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Requests prometheus.Counter
	Errors   prometheus.Counter
	Latency  prometheus.Histogram
}

func New() *Metrics {
	m := &Metrics{
		Requests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "conduit_requests_total",
		}),
		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "conduit_errors_total",
		}),
		Latency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "conduit_request_duration_seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}

	prometheus.MustRegister(m.Requests, m.Errors, m.Latency)
	return m
}
