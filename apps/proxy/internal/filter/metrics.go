package filter

import (
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/metrics"
)

// MetricsFilter records request performance data.
type MetricsFilter struct {
	store *metrics.Store
}

func NewMetricsFilter(store *metrics.Store) *MetricsFilter {
	return &MetricsFilter{store: store}
}

func (f *MetricsFilter) Before(ctx *core.Context) error {
	return nil
}

func (f *MetricsFilter) After(ctx *core.Context) error {
	latency := time.Since(ctx.StartTime)

	success := ctx.StatusCode >= 200 && ctx.StatusCode < 500

	f.store.Record(success, latency)

	return nil
}
