package balancer

import (
	"sync"

	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

type RoundRobin struct {
	mu  sync.Mutex
	idx int
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (r *RoundRobin) Pick(ups []*upstream.Upstream) *upstream.Upstream {

	if len(ups) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	n := len(ups)

	for i := 0; i < n; i++ {

		u := ups[r.idx%n]
		r.idx++

		if u.Breaker == nil || u.Breaker.Allow() {
			return u
		}
	}

	return nil
}
