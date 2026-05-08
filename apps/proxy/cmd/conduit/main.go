package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/balancer"
	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/filter"
	"github.com/alpercitak/conduit/apps/proxy/internal/metrics"
	"github.com/alpercitak/conduit/apps/proxy/internal/router"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

func main() {

	pool := upstream.NewPool([]*upstream.Upstream{
		{ID: "alpha", URL: "http://localhost:3001"},
		{ID: "beta", URL: "http://localhost:3002"},
	})

	rr := balancer.NewRoundRobin()

	r := router.New(pool, rr)

	store := metrics.NewStore()

	chain := filter.NewChain(
		filter.NewLoggingFilter(),
		filter.NewMetricsFilter(store),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		rw := core.NewResponseWriter(w)

		ctx := &core.Context{
			RequestID: core.NewRequestID(),
			StartTime: time.Now(),
			Request:   req,
		}

		if err := chain.Before(ctx); err != nil {
			http.Error(w, "request blocked by filter", 403)
			return
		}

		u, err := r.Pick(req.Context())
		if err != nil || u == nil {
			http.Error(w, "no upstream available", 503)
			return
		}

		target, _ := url.Parse(u.URL)
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(rw, req)

		ctx.StatusCode = rw.StatusCode

		_ = chain.After(ctx)
	})

	log.Println("[conduit] running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
