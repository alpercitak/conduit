package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	proxyURL := os.Getenv("PROXY_URL")
	if proxyURL == "" {
		proxyURL = "http://localhost:8080/"
	}

	traffic := 500 // base RPS

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 1000 // default 2 opens a new connection per request

	client := &http.Client{Timeout: 3 * time.Second, Transport: transport}

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
			go func(i int) {
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

				// Drain so the connection can be reused.
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()

				if lat > 500*time.Millisecond {
					fmt.Println("[SLOW]", lat)
				}
			}(i)
		}
	}
}
