package executor

import (
	"net/http"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/filter"
	"github.com/alpercitak/conduit/apps/proxy/internal/metrics"
	"github.com/alpercitak/conduit/apps/proxy/internal/router"
)

type Executor struct {
	router  *router.SmartRouter
	tracer  core.Tracer
	filters *filter.Chain
	metrics *metrics.Metrics
}

func New(
	router *router.SmartRouter,
	tracer core.Tracer,
	filters *filter.Chain,
	m *metrics.Metrics,
) *Executor {
	return &Executor{
		router:  router,
		tracer:  tracer,
		filters: filters,
		metrics: m,
	}
}

func (e *Executor) finalize(ctx *core.Context) {
	e.tracer.Mark(ctx, "request_end")
}

func (e *Executor) Handle(w http.ResponseWriter, req *http.Request) {
	rw := core.NewResponseWriter(w)

	ctx := &core.Context{
		RequestID: core.NewRequestID(),
		StartTime: time.Now(),
		Request:   req,
		State:     make(core.State),
	}

	defer e.finalize(ctx)

	e.tracer.Mark(ctx, "request_start")

	err := e.filters.Execute(ctx, func(ctx *core.Context) error {
		upstream, err := e.router.Pick()
		if err != nil {
			http.Error(w, "no upstream available", http.StatusServiceUnavailable)
			ctx.Outcome = core.OutcomeUpstreamError
			return nil
		}

		upstreamId := upstream.ID

		ctx.UpstreamID = upstreamId
		ctx.UpstreamURL = upstream.URL

		start := time.Now()

		e.tracer.Track(ctx, "proxy", func() {
			upstream.Proxy.ServeHTTP(rw, req)
		})

		latency := time.Since(start)

		ctx.StatusCode = rw.StatusCode

		e.metrics.Requests.WithLabelValues(upstreamId).Inc()
		e.metrics.Latency.WithLabelValues(upstreamId).Observe(latency.Seconds())

		if rw.StatusCode >= 500 {
			e.metrics.Errors.WithLabelValues(upstreamId).Inc()
			ctx.Outcome = core.OutcomeUpstreamError
		} else if rw.StatusCode >= 400 {
			ctx.Outcome = core.OutcomeClientError
		} else {
			ctx.Outcome = core.OutcomeSuccess
		}

		return nil
	})

	if err != nil {
		http.Error(w, "blocked", http.StatusForbidden)
		ctx.Outcome = core.OutcomeBlocked
		return
	}
}
