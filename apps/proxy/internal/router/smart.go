package router

import (
	"fmt"
	"math/rand"

	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

const explorationRate = 0.3

type SmartRouter struct {
	pool *upstream.Pool
}

func NewSmart(pool *upstream.Pool) *SmartRouter {
	return &SmartRouter{
		pool: pool,
	}
}

func (r *SmartRouter) Pick() (*upstream.Upstream, error) {
	healthy := r.healthy()

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy upstreams")
	}

	// exploration
	if rand.Float64() < explorationRate {
		return healthy[rand.Intn(len(healthy))], nil
	}

	var best *upstream.Upstream
	bestScore := -1.0

	for _, u := range healthy {
		score := Score(u)

		if score > bestScore {
			best = u
			bestScore = score
		}
	}

	return best, nil
}

func (r *SmartRouter) healthy() []*upstream.Upstream {
	var healthy []*upstream.Upstream

	for _, u := range r.pool.All() {
		if u.Breaker != nil && !u.Breaker.Allow() {
			continue
		}

		healthy = append(healthy, u)
	}

	return healthy
}

func Score(u *upstream.Upstream) float64 {
	if u == nil {
		return -1
	}
	return 1000 - u.LatencyEWMA - (u.ErrorEWMA * 500)
}
