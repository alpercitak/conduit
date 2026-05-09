package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/circuit"
	"github.com/alpercitak/conduit/apps/proxy/internal/config"
	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/executor"
	"github.com/alpercitak/conduit/apps/proxy/internal/filter"
	"github.com/alpercitak/conduit/apps/proxy/internal/metrics"
	"github.com/alpercitak/conduit/apps/proxy/internal/router"
	"github.com/alpercitak/conduit/apps/proxy/internal/tracer"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	inMemoryTracer := core.NewInMemoryTracer()

	alphaURL := config.GetEnv("ALPHA_URL", "http://localhost:3001")
	betaURL := config.GetEnv("BETA_URL", "http://localhost:3002")

	alpha, err := upstream.New("alpha", alphaURL, circuit.New(5, 10*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	beta, err := upstream.New("beta", betaURL, circuit.New(5, 10*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	upstreamPool := upstream.NewPool([]*upstream.Upstream{alpha, beta})
	smartRouter := router.NewSmart(upstreamPool)
	metrics := metrics.New()
	filterChain := filter.NewChain(
		filter.NewLoggingFilter(),
		filter.NewResilienceFilter(func(ctx *core.Context) *upstream.Upstream {
			return upstreamPool.GetByID(ctx.UpstreamID)
		}),
	)
	exec := executor.New(smartRouter, inMemoryTracer, filterChain, metrics)

	http.HandleFunc("/", exec.Handle)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/trace", tracer.HTTPHandler(inMemoryTracer))

	log.Println("[conduit] running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
