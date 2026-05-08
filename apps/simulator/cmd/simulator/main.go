package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	StatusCode int
	Latency    time.Duration
	Err        error
}

func hit(client *http.Client, url string) Result {
	start := time.Now()

	resp, err := client.Get(url)
	if err != nil {
		return Result{Err: err}
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	return Result{
		StatusCode: resp.StatusCode,
		Latency:    time.Since(start),
	}
}

func main() {
	target := "http://localhost:8080/"

	const (
		totalRequests = 200
		concurrency   = 20
	)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	results := make(chan Result, totalRequests)

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	fmt.Println("[simulator] starting load test...")
	fmt.Printf("[simulator] target=%s requests=%d concurrency=%d\n",
		target, totalRequests, concurrency)

	startAll := time.Now()

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			results <- hit(client, target)
		}()
	}

	wg.Wait()
	close(results)

	duration := time.Since(startAll)

	var (
		ok, errCount int
		totalLatency time.Duration
		maxLatency   time.Duration
		minLatency   = time.Hour
	)

	for r := range results {
		if r.Err != nil {
			errCount++
			continue
		}

		ok++
		totalLatency += r.Latency

		if r.Latency > maxLatency {
			maxLatency = r.Latency
		}
		if r.Latency < minLatency {
			minLatency = r.Latency
		}
	}

	avg := time.Duration(0)
	if ok > 0 {
		avg = totalLatency / time.Duration(ok)
	}

	fmt.Println("\n--- simulation results ---")
	fmt.Printf("total time:   %s\n", duration)
	fmt.Printf("requests:     %d\n", totalRequests)
	fmt.Printf("success:      %d\n", ok)
	fmt.Printf("errors:       %d\n", errCount)
	fmt.Printf("avg latency:  %s\n", avg)
	fmt.Printf("min latency:  %s\n", minLatency)
	fmt.Printf("max latency:  %s\n", maxLatency)
	fmt.Println("--------------------------")
}
