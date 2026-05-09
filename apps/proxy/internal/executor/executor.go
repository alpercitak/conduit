package executor

import (
	"net/http"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/filter"
	"github.com/alpercitak/conduit/apps/proxy/internal/router"
)

type Executor struct {
	router  *router.SmartRouter
	tracer  core.Tracer
	filters *filter.Chain
}

func New(
	router *router.SmartRouter,
	tracer core.Tracer,
	filters *filter.Chain,
) *Executor {
	return &Executor{
		router:  router,
		tracer:  tracer,
		filters: filters,
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

		u, err := e.router.Pick()
		if err != nil {
			http.Error(w, "no upstream available", http.StatusServiceUnavailable)

			ctx.Outcome = core.OutcomeUpstreamError

			return nil
		}

		ctx.UpstreamID = u.ID
		ctx.UpstreamURL = u.URL

		start := time.Now()

		e.tracer.Track(ctx, "proxy", func() {
			u.Proxy.ServeHTTP(rw, req)
		})

		latency := time.Since(start)

		u.RecordLatency(latency)

		ctx.StatusCode = rw.StatusCode

		switch {
		case rw.StatusCode >= 500:
			u.RecordFailure()
			ctx.Outcome = core.OutcomeUpstreamError

		case rw.StatusCode >= 400:
			u.RecordFailure()
			ctx.Outcome = core.OutcomeClientError

		default:
			u.RecordSuccess()
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
