package upstream

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"proxy/internal/circuit"
)

const (
	alphaLatency = 0.2
	alphaError   = 0.3

	// StatusClientClosedRequest (nginx's 499) marks requests the client
	// abandoned. It is below 500, so it doesn't count against the upstream.
	StatusClientClosedRequest = 499
)

type Upstream struct {
	ID      string
	Proxy   *httputil.ReverseProxy
	Breaker *circuit.Breaker

	mu          sync.RWMutex
	latencyEWMA float64 // milliseconds
	errorEWMA   float64 // 0..1
	hasLatency  bool
}

// New proxies to rawURL. timeout caps how long a hung upstream can hold a
// request before it fails with 502; zero means no limit.
func New(id, rawURL string, breaker *circuit.Breaker, timeout time.Duration) (*Upstream, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = timeout
	transport.MaxIdleConnsPerHost = 100 // default 2 forces a new connection per request under load

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	proxy.ErrorHandler = proxyError

	return &Upstream{
		ID:      id,
		Proxy:   proxy,
		Breaker: breaker,
	}, nil
}

func proxyError(w http.ResponseWriter, r *http.Request, err error) {
	if r.Context().Err() != nil {
		w.WriteHeader(StatusClientClosedRequest)
		return
	}
	log.Printf("proxy error: %v", err)
	w.WriteHeader(http.StatusBadGateway)
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
}

// Stats returns the current latency EWMA (ms) and error-rate EWMA.
func (u *Upstream) Stats() (latencyMs, errorRate float64) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.latencyEWMA, u.errorEWMA
}
