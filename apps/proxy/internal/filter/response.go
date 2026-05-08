package filter

import (
	"github.com/alpercitak/conduit/apps/proxy/internal/core"
	"github.com/alpercitak/conduit/apps/proxy/internal/upstream"
)

type ResponseFilter struct {
	Upstream *upstream.Upstream
}

func NewResponseFilter(u *upstream.Upstream) *ResponseFilter {
	return &ResponseFilter{Upstream: u}
}

func (f *ResponseFilter) Before(ctx *core.Context) error {
	return nil
}

func (f *ResponseFilter) After(ctx *core.Context) error {
	if f.Upstream == nil || f.Upstream.Breaker == nil {
		return nil
	}

	if ctx.StatusCode >= 500 {
		f.Upstream.Breaker.Failure()
	} else {
		f.Upstream.Breaker.Success()
	}

	return nil
}
