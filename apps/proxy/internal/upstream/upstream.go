package upstream

import (
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/circuit"
)

const (
	alphaLatency  = 0.2
	alphaError    = 0.3
	decayHalfLife = 30 * time.Second
)

type Upstream struct {
	ID      string
	URL     string
	Breaker *circuit.Breaker

	LatencyEWMA float64
	ErrorEWMA   float64

	lastUpdate time.Time
}

func (u *Upstream) UpdateLatency(d time.Duration) {
	lat := float64(d.Milliseconds())

	if u.LatencyEWMA == 0 {
		u.LatencyEWMA = lat
	} else {
		u.LatencyEWMA = alphaLatency*lat + (1-alphaLatency)*u.LatencyEWMA
	}

	u.lastUpdate = time.Now()
}

func (u *Upstream) RecordSuccess() {
	u.updateError(false)
}

func (u *Upstream) RecordFailure() {
	u.updateError(true)
}

func (u *Upstream) updateError(failed bool) {
	var errVal float64
	if failed {
		errVal = 1
	} else {
		errVal = 0
	}

	if u.ErrorEWMA == 0 {
		u.ErrorEWMA = errVal
	} else {
		u.ErrorEWMA = alphaError*errVal + (1-alphaError)*u.ErrorEWMA
	}

	u.lastUpdate = time.Now()
}

func (u *Upstream) DecayedLatency() float64 {
	return u.LatencyEWMA * decayFactor(u.lastUpdate)
}

func (u *Upstream) DecayedError() float64 {
	return u.ErrorEWMA * decayFactor(u.lastUpdate)
}

func decayFactor(t time.Time) float64 {
	age := time.Since(t).Seconds()
	half := decayHalfLife.Seconds()

	return 0.5 * (age / half)
}

func (u *Upstream) Observe(statusCode int, latency time.Duration) {
	u.UpdateLatency(latency)

	if statusCode >= 500 {
		u.RecordFailure()
		return
	}

	if statusCode >= 400 {
		u.RecordFailure()
		return
	}

	u.RecordSuccess()
}
