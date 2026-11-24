package upstream

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"proxy/internal/circuit"
)

const (
	alphaLatency = 0.2
	alphaError   = 0.3
)

type Upstream struct {
	ID      string
	URL     string
	Proxy   *httputil.ReverseProxy
	Breaker *circuit.Breaker

	mu          sync.RWMutex
	latencyEWMA float64 // milliseconds
	errorEWMA   float64 // 0..1
	hasLatency  bool
	lastUpdate  time.Time
}

func New(id, rawURL string, breaker *circuit.Breaker) (*Upstream, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &Upstream{
		ID:      id,
		URL:     rawURL,
		Proxy:   httputil.NewSingleHostReverseProxy(target),
		Breaker: breaker,
	}, nil
}

// Record folds one request's latency and outcome into the EWMAs.
func (u *Upstream) Record(d time.Duration, failed bool) {
	lat := float64(d) / float64(time.Millisecond)

	errVal := 0.0
	if failed {
		errVal = 1
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.hasLatency {
		u.latencyEWMA = lat
		u.hasLatency = true
	} else {
		u.latencyEWMA = alphaLatency*lat + (1-alphaLatency)*u.latencyEWMA
	}

	// Error rate starts at a prior of 0 (healthy), so a single failure
	// moves it by alphaError instead of jumping straight to 1.
	u.errorEWMA = alphaError*errVal + (1-alphaError)*u.errorEWMA

	u.lastUpdate = time.Now()
}

// Stats returns the current latency EWMA (ms) and error-rate EWMA.
func (u *Upstream) Stats() (latencyMs, errorRate float64) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.latencyEWMA, u.errorEWMA
}
