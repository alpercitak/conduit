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

	u := ups[r.idx%len(ups)]
	r.idx++

	return u
}
