package filter

import (
	"log"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
)

type MetricsFilter struct{}

func NewMetricsFilter() *MetricsFilter {
	return &MetricsFilter{}
}

func (f *MetricsFilter) Before(ctx *core.Context) error {
	return nil
}

func (f *MetricsFilter) After(ctx *core.Context) error {

	duration := time.Since(ctx.StartTime)

	log.Printf(
		"[metrics] req=%s duration=%s status=%d",
		ctx.RequestID,
		duration,
		ctx.StatusCode,
	)

	return nil
}
