package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Requests *prometheus.CounterVec
	Errors   *prometheus.CounterVec
	Latency  *prometheus.HistogramVec
}

func New() *Metrics {
	m := &Metrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "conduit_requests_total",
				Help: "Total number of requests processed",
			},
			[]string{"upstream"},
		),

		Errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "conduit_errors_total",
				Help: "Total number of failed requests",
			},
			[]string{"upstream"},
		),

		Latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "conduit_request_duration_seconds",
				Help:    "Request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"upstream"},
		),
	}

	prometheus.MustRegister(
		m.Requests,
		m.Errors,
		m.Latency,
	)

	return m
}
