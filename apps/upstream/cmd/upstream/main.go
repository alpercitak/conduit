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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d := latency
		if jitter > 0 {
			d += time.Duration(rand.Int63n(int64(jitter)))
		}
		time.Sleep(d)

		if rand.Float64() < errorRate {
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
