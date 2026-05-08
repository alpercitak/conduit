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

	success int64
	failure int64
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
	u.success++
	u.recalculateErrorRate()
}

func (u *Upstream) RecordFailure() {
	u.failure++
	u.recalculateErrorRate()
}

func (u *Upstream) recalculateErrorRate() {
	total := u.success + u.failure
	if total == 0 {
		u.ErrorRate = 0
		return
	}

	u.ErrorRate = float64(u.failure) / float64(total)
}
