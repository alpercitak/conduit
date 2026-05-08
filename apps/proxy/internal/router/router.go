package router

import (
	"context"

	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

type Router struct {
	pool *upstream.Pool
}

func New(pool *upstream.Pool) *Router {
	return &Router{pool: pool}
}

func (r *Router) Pick(ctx context.Context) (*upstream.Upstream, error) {
	return r.pool.First()
}
