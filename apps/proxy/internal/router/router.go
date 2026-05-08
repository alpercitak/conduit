package router

import (
	"context"

	"github.com/alpercitak/conduit/apps/proxy/internal/balancer"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

type Router struct {
	pool     *upstream.Pool
	balancer balancer.Balancer
}

func New(pool *upstream.Pool, b balancer.Balancer) *Router {
	return &Router{
		pool:     pool,
		balancer: b,
	}
}

func (r *Router) Pick(ctx context.Context) (*upstream.Upstream, error) {

	ups := r.pool.All()

	u := r.balancer.Pick(ups)
	if u == nil {
		return nil, nil
	}

	return u, nil
}
