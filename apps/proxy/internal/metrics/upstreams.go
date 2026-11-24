package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"proxy/internal/circuit"
	"proxy/internal/upstream"
)

var breakerStateValue = map[circuit.State]float64{
	circuit.Closed:   0,
	circuit.HalfOpen: 1,
	circuit.Open:     2,
}

// RegisterUpstreams exports each upstream's routing inputs, read at scrape
// time: breaker state (0 closed, 1 half-open, 2 open), latency and error
// EWMAs, and the router's score.
func RegisterUpstreams(ups []*upstream.Upstream, score func(*upstream.Upstream) float64) {
	for _, u := range ups {
		prometheus.MustRegister(upstreamGauges(u, score)...)
	}
}

func upstreamGauges(u *upstream.Upstream, score func(*upstream.Upstream) float64) []prometheus.Collector {
	gauge := func(name, help string, fn func() float64) prometheus.Collector {
		return prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name:        name,
			Help:        help,
			ConstLabels: prometheus.Labels{"upstream": u.ID},
		}, fn)
	}

	gauges := []prometheus.Collector{
		gauge("conduit_upstream_latency_ewma_ms", "Latency EWMA in milliseconds", func() float64 {
			latency, _ := u.Stats()
			return latency
		}),
		gauge("conduit_upstream_error_ewma", "Error rate EWMA (0..1)", func() float64 {
			_, errRate := u.Stats()
			return errRate
		}),
		gauge("conduit_upstream_score", "Router score (higher wins)", func() float64 {
			return score(u)
		}),
	}

	if u.Breaker != nil {
		gauges = append(gauges, gauge("conduit_breaker_state", "Breaker state: 0 closed, 1 half-open, 2 open", func() float64 {
			return breakerStateValue[u.Breaker.State()]
		}))
	}

	return gauges
}
