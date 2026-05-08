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

		err := chain.Execute(ctx, func(ctx *core.Context) error {

			u, err := r.Pick(req.Context())
			if err != nil || u == nil {
				http.Error(w, "no upstream available", 503)
				return nil
			}

			ctx.UpstreamID = u.ID
			ctx.UpstreamURL = u.URL

			target, _ := url.Parse(u.URL)
			proxy := httputil.NewSingleHostReverseProxy(target)

			proxy.ServeHTTP(rw, req)

			ctx.StatusCode = rw.StatusCode

			return nil
		})

		if err != nil {
			http.Error(w, "filter chain blocked request", 403)
			return
		}
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

	log.Println("[conduit] running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
