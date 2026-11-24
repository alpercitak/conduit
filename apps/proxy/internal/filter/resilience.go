package filter

import (
	"proxy/internal/core"
	"proxy/internal/upstream"
)

type ResilienceFilter struct {
	UpstreamResolver func(ctx *core.Context) *upstream.Upstream
}

func NewResilienceFilter(resolver func(ctx *core.Context) *upstream.Upstream) *ResilienceFilter {
	return &ResilienceFilter{UpstreamResolver: resolver}
}

func (f *ResilienceFilter) Before(ctx *core.Context) error {
	return nil
}

func (f *ResilienceFilter) After(ctx *core.Context) error {
	u := f.UpstreamResolver(ctx)
	if u == nil || u.Breaker == nil || ctx.StatusCode == upstream.StatusClientClosedRequest {
		return nil
	}

	if ctx.StatusCode >= 500 {
		u.Breaker.Failure()
	} else {
		u.Breaker.Success()
	}

	return nil
}
