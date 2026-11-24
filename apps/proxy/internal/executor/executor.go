package executor

import (
	"net/http"
	"time"

	"proxy/internal/core"
	"proxy/internal/filter"
	"proxy/internal/metrics"
	"proxy/internal/router"
	"proxy/internal/upstream"
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
	}

	defer e.finalize(ctx)

	e.tracer.Mark(ctx, "request_start")

	err := e.filters.Execute(ctx, func(ctx *core.Context) error {
		u, err := e.router.Pick()
		if err != nil {
			http.Error(w, "no upstream available", http.StatusServiceUnavailable)
			ctx.Outcome = core.OutcomeUpstreamError
			return nil
		}

		upstreamId := u.ID

		ctx.UpstreamID = upstreamId

		start := time.Now()

		e.tracer.Track(ctx, "proxy", func() {
			u.Proxy.ServeHTTP(rw, req)
		})

		latency := time.Since(start)

		ctx.StatusCode = rw.StatusCode

		// An abandoned request says nothing about the upstream's health.
		if rw.StatusCode != upstream.StatusClientClosedRequest {
			u.Record(latency, rw.StatusCode >= 500)
		}

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
