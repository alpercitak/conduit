package router

import (
	"fmt"
	"math/rand"
	"slices"

	"proxy/internal/upstream"
)

const explorationRate = 0.3

type SmartRouter struct {
	pool        *upstream.Pool
	exploration float64
}

func NewSmart(pool *upstream.Pool) *SmartRouter {
	return &SmartRouter{
		pool:        pool,
		exploration: explorationRate,
	}
}

func (r *SmartRouter) Pick() (*upstream.Upstream, error) {
	candidates := r.available()

	for len(candidates) > 0 {
		i := r.choose(candidates)
		u := candidates[i]

		// Claim the breaker only for the upstream we actually send to.
		// It can still refuse if a concurrent request took the half-open probe.
		if u.Breaker == nil || u.Breaker.Allow() {
			return u, nil
		}

		candidates = slices.Delete(candidates, i, i+1)
	}

	return nil, fmt.Errorf("no healthy upstreams")
}

func (r *SmartRouter) choose(candidates []*upstream.Upstream) int {
	if rand.Float64() < r.exploration {
		return rand.Intn(len(candidates))
	}

	best := 0
	bestScore := Score(candidates[0])

	for i, u := range candidates[1:] {
		if s := Score(u); s > bestScore {
			best, bestScore = i+1, s
		}
	}

	return best
}

func (r *SmartRouter) available() []*upstream.Upstream {
	var out []*upstream.Upstream

	for _, u := range r.pool.All() {
		if u.Breaker != nil && !u.Breaker.Available() {
			continue
		}
		out = append(out, u)
	}

	return out
}

func Score(u *upstream.Upstream) float64 {
	latency, errRate := u.Stats()
	return 1000 - latency - (errRate * 500)
}
