package filter

import (
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/metrics"
)

type MetricsFilter struct {
	m *metrics.Metrics
}

func NewMetricsFilter(m *metrics.Metrics) *MetricsFilter {
	return &MetricsFilter{m: m}
}

func (f *MetricsFilter) Before(ctx *core.Context) error {
	return nil
}

func (f *MetricsFilter) After(ctx *core.Context) error {
	f.m.Requests.Inc()

	if ctx.StatusCode >= 500 {
		f.m.Errors.Inc()
	}

	f.m.Latency.Observe(time.Since(ctx.StartTime).Seconds())

	return nil
}
