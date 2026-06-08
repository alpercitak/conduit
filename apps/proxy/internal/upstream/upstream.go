package upstream

import (
	"net/http/httputil"
	"net/url"
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

	LatencyEWMA float64
	ErrorEWMA   float64

	LastUpdate time.Time
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

func (u *Upstream) RecordLatency(d time.Duration) {
	lat := float64(d.Milliseconds())

	if u.LatencyEWMA == 0 {
		u.LatencyEWMA = lat
	} else {
		u.LatencyEWMA =
			alphaLatency*lat +
				(1-alphaLatency)*u.LatencyEWMA
	}

	u.LastUpdate = time.Now()
}

func (u *Upstream) RecordSuccess() {
	u.recordError(0)
}

func (u *Upstream) RecordFailure() {
	u.recordError(1)
}

func (u *Upstream) recordError(v float64) {
	if u.ErrorEWMA == 0 {
		u.ErrorEWMA = v
	} else {
		u.ErrorEWMA =
			alphaError*v +
				(1-alphaError)*u.ErrorEWMA
	}

	u.LastUpdate = time.Now()
}
