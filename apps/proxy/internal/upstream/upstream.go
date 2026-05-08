package upstream

import (
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/circuit"
)

type Upstream struct {
	ID      string
	URL     string
	Breaker *circuit.Breaker

	LatencyAvg time.Duration
	ErrorRate  float64
}

func (u *Upstream) UpdateLatency(d time.Duration) {
	const alpha = 0.2

	if u.LatencyAvg == 0 {
		u.LatencyAvg = d
		return
	}

	u.LatencyAvg = time.Duration(
		float64(u.LatencyAvg)*(1-alpha) + float64(d)*alpha,
	)
}

func (u *Upstream) RecordSuccess() {
	const alpha = 0.1
	u.ErrorRate = u.ErrorRate * (1 - alpha)
}

func (u *Upstream) RecordFailure() {
	const alpha = 0.1
	u.ErrorRate = u.ErrorRate*(1-alpha) + 1*alpha
}
