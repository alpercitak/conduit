package executor

import (
	"net/http"
	"net/http/httputil"
	"net/url"
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

func New(router *router.SmartRouter, tracer core.Tracer, filters *filter.Chain) *Executor {
	return &Executor{
		router:  router,
		tracer:  tracer,
		filters: filters,
	}
}

func (e *Executor) finalize(ctx *core.Context) {
	e.tracer.Mark(ctx, "request_end")

	if ctx.Outcome == "" {
		switch {
		case ctx.StatusCode >= 500:
			ctx.Outcome = core.OutcomeServerError
		case ctx.StatusCode >= 400:
			ctx.Outcome = core.OutcomeClientError
		default:
			ctx.Outcome = core.OutcomeSuccess
		}
	}

	e.tracer.Mark(ctx, "request_finalized")
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

		u, err := e.router.Pick(ctx)
		if err != nil || u == nil {
			http.Error(w, "no upstream available", 503)
			ctx.Outcome = core.OutcomeUpstreamError
			return nil
		}

		ctx.UpstreamID = u.ID
		ctx.UpstreamURL = u.URL

		target, err := url.Parse(u.URL)
		if err != nil {
			http.Error(w, "bad upstream url", 500)
			ctx.Outcome = core.OutcomeServerError
			return nil
		}

		proxy := httputil.NewSingleHostReverseProxy(target)

		start := time.Now()

		var success bool = true

		e.tracer.Track(ctx, "proxy", func() {
			proxy.ServeHTTP(rw, req)
		})

		if rw.StatusCode >= 500 {
			success = false
		}

		u.Observe(rw.StatusCode, time.Since(start))

		if success {
			u.RecordSuccess()
		} else {
			u.RecordFailure()
		}

		ctx.StatusCode = rw.StatusCode

		return nil
	})

	if err != nil {
		http.Error(w, "blocked", 403)
		ctx.Outcome = core.OutcomeBlocked
		return
	}
}
