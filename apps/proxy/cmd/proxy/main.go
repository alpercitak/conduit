package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"proxy/internal/circuit"
	"proxy/internal/config"
	"proxy/internal/core"
	"proxy/internal/executor"
	"proxy/internal/filter"
	"proxy/internal/metrics"
	"proxy/internal/router"
	"proxy/internal/tracer"
	"proxy/internal/upstream"
)

func main() {
	inMemoryTracer := core.NewInMemoryTracer()

	alphaURL := config.GetEnv("ALPHA_URL", "http://localhost:3001")
	betaURL := config.GetEnv("BETA_URL", "http://localhost:3002")

	errorThreshold := config.GetFloat("BREAKER_ERROR_THRESHOLD", 0.5)
	retryTimeout := config.GetDuration("BREAKER_RETRY_TIMEOUT", 10*time.Second)
	exploration := config.GetFloat("EXPLORATION_RATE", 0.3)
	timeout := config.GetDuration("UPSTREAM_TIMEOUT", 2*time.Second)

	log.Printf("[conduit] breaker threshold=%.2f retry=%s, exploration=%.2f, upstream timeout=%s",
		errorThreshold, retryTimeout, exploration, timeout)

	alpha, err := upstream.New("alpha", alphaURL, circuit.New(errorThreshold, retryTimeout), timeout)
	if err != nil {
		log.Fatal(err)
	}

	beta, err := upstream.New("beta", betaURL, circuit.New(errorThreshold, retryTimeout), timeout)
	if err != nil {
		log.Fatal(err)
	}

	upstreamPool := upstream.NewPool([]*upstream.Upstream{alpha, beta})
	smartRouter := router.NewSmart(upstreamPool, exploration)
	proxyMetrics := metrics.New()
	metrics.RegisterUpstreams(upstreamPool.All(), router.Score)
	filterChain := filter.NewChain(
		filter.NewLoggingFilter(),
		filter.NewResilienceFilter(func(ctx *core.Context) *upstream.Upstream {
			return upstreamPool.GetByID(ctx.UpstreamID)
		}),
	)
	exec := executor.New(smartRouter, inMemoryTracer, filterChain, proxyMetrics)

	http.HandleFunc("/", exec.Handle)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/trace", tracer.HTTPHandler(inMemoryTracer))

	addr := ":" + config.GetEnv("PORT", "8080")
	log.Println("[conduit] running on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
