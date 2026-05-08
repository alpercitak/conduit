package router

import (
	"fmt"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

type SmartRouter struct {
	pool *upstream.Pool
}

func NewSmart(pool *upstream.Pool) *SmartRouter {
	return &SmartRouter{pool: pool}
}

func (r *SmartRouter) Pick(ctx *core.Context) (*upstream.Upstream, error) {

	var best *upstream.Upstream
	bestScore := -1.0

	for _, u := range r.pool.All() {

		if u.Breaker != nil && !u.Breaker.Allow() {
			continue
		}

		score := Score(u)

		if score > bestScore {
			best = u
			bestScore = score
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no healthy upstream available")
	}

	return best, nil
}

func Score(u *upstream.Upstream) float64 {
	if u == nil {
		return -1
	}

	lat := u.DecayedLatency()
	err := u.DecayedError() * 1000

	return 1000 - lat - err
}
