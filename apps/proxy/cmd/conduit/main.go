package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/circuit"
	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/executor"
	"github.com/alpercitak/conduit/apps/proxy/internal/filter"
	"github.com/alpercitak/conduit/apps/proxy/internal/metrics"
	"github.com/alpercitak/conduit/apps/proxy/internal/router"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

func main() {

	tracer := core.NewInMemoryTracer()

	pool := upstream.NewPool([]*upstream.Upstream{
		{ID: "alpha", URL: "http://localhost:3001", Breaker: circuit.New(5, 10*time.Second)},
		{ID: "beta", URL: "http://localhost:3002", Breaker: circuit.New(5, 10*time.Second)},
	})

	r := router.NewSmart(pool)

	store := metrics.NewStore()

	chain := filter.NewChain(
		filter.NewLoggingFilter(),
		filter.NewMetricsFilter(store),
		filter.NewResilienceFilter(func(ctx *core.Context) *upstream.Upstream {
			return pool.GetByID(ctx.UpstreamID)
		}),
	)

	exec := executor.New(r, tracer, chain)

	http.HandleFunc("/", exec.Handle)

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

		trace, ok := tracer.Get(id)
		if !ok {
			http.Error(w, "trace not found", 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trace)
	})

	log.Println("[conduit] running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
