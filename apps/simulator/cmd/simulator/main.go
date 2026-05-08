package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Profile struct {
	Name          string
	URL           string
	LatencyBase   time.Duration
	LatencyJitter time.Duration
	ErrorRate     float64
}

func (p Profile) Do(req *http.Request) (*http.Response, error) {
	// simulate latency
	lat := p.LatencyBase + time.Duration(rand.Int63n(int64(p.LatencyJitter)))
	time.Sleep(lat)

	// simulate failure
	if rand.Float64() < p.ErrorRate {
		return nil, fmt.Errorf("simulated upstream failure: %s", p.Name)
	}

	return &http.Response{
		StatusCode: 200,
		Body:       http.NoBody,
	}, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	proxyURL := "http://localhost:8080/"

	traffic := 500 // base RPS

	// alpha := Profile{
	// 	Name:          "alpha",
	// 	LatencyBase:   30 * time.Millisecond,
	// 	LatencyJitter: 50 * time.Millisecond,
	// 	ErrorRate:     0.01,
	// }

	// beta := Profile{
	// 	Name:          "beta",
	// 	LatencyBase:   200 * time.Millisecond,
	// 	LatencyJitter: 800 * time.Millisecond,
	// 	ErrorRate:     0.15,
	// }

	client := &http.Client{Timeout: 3 * time.Second}

	var wg sync.WaitGroup

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {

		// BURST MODE
		currentRPS := traffic
		if rand.Float64() < 0.2 {
			currentRPS = traffic * 8 // spikes
			fmt.Println("[SIM] BURST x8")
		}

		for i := 0; i < currentRPS; i++ {
			wg.Add(1)

			go func(i int) {
				defer wg.Done()

				req, _ := http.NewRequest("GET", proxyURL, nil)

				// just random noise header
				req.Header.Set("X-Sim-Req", fmt.Sprint(i))

				start := time.Now()
				resp, err := client.Do(req)
				lat := time.Since(start)

				if err != nil {
					fmt.Println("[SIM ERROR]", err, "lat:", lat)
					return
				}

				_ = resp.Body.Close()

				if lat > 500*time.Millisecond {
					fmt.Println("[SLOW]", lat)
				}
			}(i)
		}

		wg.Wait()
	}
}
