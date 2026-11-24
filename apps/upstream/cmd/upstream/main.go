package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	name := os.Getenv("UPSTREAM_NAME")
	if name == "" {
		name = "upstream"
	}

	// Injected behavior: each response waits LATENCY_MS plus up to JITTER_MS,
	// then fails with 500 at ERROR_RATE (0..1).
	latency := time.Duration(envFloat("LATENCY_MS", 0)) * time.Millisecond
	jitter := time.Duration(envFloat("JITTER_MS", 0)) * time.Millisecond
	errorRate := envFloat("ERROR_RATE", 0)

	// Optional degraded phase: for the first DEGRADE_FOR of every
	// DEGRADE_EVERY, add DEGRADED_LATENCY_MS and fail at DEGRADED_ERROR_RATE.
	degradeEvery := envDuration("DEGRADE_EVERY")
	degradeFor := envDuration("DEGRADE_FOR")
	degradedLatency := time.Duration(envFloat("DEGRADED_LATENCY_MS", 0)) * time.Millisecond
	degradedErrorRate := envFloat("DEGRADED_ERROR_RATE", 0)
	start := time.Now()

	degraded := func() bool {
		return degradeEvery > 0 && time.Since(start)%degradeEvery < degradeFor
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d, failRate := latency, errorRate
		if degraded() {
			d, failRate = d+degradedLatency, degradedErrorRate
		}
		if jitter > 0 {
			d += time.Duration(rand.Int63n(int64(jitter)))
		}
		time.Sleep(d)

		if rand.Float64() < failRate {
			http.Error(w, "simulated failure from "+name, http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "response from %s\n", name)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("[%s] listening on :%s (latency=%s jitter=%s errors=%.2f)", name, port, latency, jitter, errorRate)
	if degradeEvery > 0 {
		log.Printf("[%s] degrades for %s every %s (+%s, errors=%.2f)", name, degradeFor, degradeEvery, degradedLatency, degradedErrorRate)
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func envFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		log.Fatalf("invalid %s=%q: %v", key, v, err)
	}
	return f
}

func envDuration(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("invalid %s=%q: %v", key, v, err)
	}
	return d
}
