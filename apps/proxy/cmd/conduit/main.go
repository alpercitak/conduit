package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/balancer"
	"github.com/alpercitak/conduit/apps/proxy/internal/circuit"
	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/filter"
	"github.com/alpercitak/conduit/apps/proxy/internal/metrics"
	"github.com/alpercitak/conduit/apps/proxy/internal/router"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

func main() {

	var traceStore = make(map[string][]core.TraceEvent)

	pool := upstream.NewPool([]*upstream.Upstream{
		{ID: "alpha", URL: "http://localhost:3001", Breaker: circuit.New(5, 10*time.Second)},
		{ID: "beta", URL: "http://localhost:3002", Breaker: circuit.New(5, 10*time.Second)},
	})

	rr := balancer.NewRoundRobin()

	r := router.New(pool, rr)

	store := metrics.NewStore()

	chain := filter.NewChain(
		filter.NewLoggingFilter(),
		filter.NewMetricsFilter(store),
		filter.NewResilienceFilter(func(ctx *core.Context) *upstream.Upstream {
			return pool.GetByID(ctx.UpstreamID)
		}),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		rw := core.NewResponseWriter(w)

		ctx := &core.Context{
			RequestID: core.NewRequestID(),
			StartTime: time.Now(),
			Request:   req,
			State:     make(core.State),
		}

		ctx.Mark("request_start")

		err := chain.Execute(ctx, func(ctx *core.Context) error {
			ctx.Mark("before_routing")

			u, err := r.Pick(req.Context())
			if err != nil || u == nil {
				http.Error(w, "no upstream available", 503)
				return nil
			}

			ctx.Mark("upstream_selected")

			ctx.UpstreamID = u.ID
			ctx.UpstreamURL = u.URL

			target, _ := url.Parse(u.URL)
			proxy := httputil.NewSingleHostReverseProxy(target)

			proxy.ServeHTTP(rw, req)

			ctx.Mark("proxy_complete")

			ctx.StatusCode = rw.StatusCode

			return nil
		})

		if err != nil {
			http.Error(w, "filter chain blocked request", 403)
			return
		}

		ctx.Mark("request_end")

		traceStore[ctx.RequestID] = ctx.Trace
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {

		p50, p95, p99 := store.Summary()

		w.Header().Set("Content-Type", "text/plain")

		w.Write([]byte(
			fmt.Sprintf(
				"requests_total %d\nerrors_total %d\np50 %s\np95 %s\np99 %s\n",
				store.Requests,
				store.Errors,
				p50,
				p95,
				p99,
			),
		))
	})

	http.HandleFunc("/trace", func(w http.ResponseWriter, req *http.Request) {

		id := req.URL.Query().Get("id")

		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}

		trace, ok := traceStore[id]
		if !ok {
			http.Error(w, "trace not found", 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%v", trace)
	})

	log.Println("[conduit] running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
