package filter

import "github.com/alpercitak/conduit/apps/proxy/internal/core"

type Chain struct {
	filters []Filter
}

type Handler func(ctx *core.Context) error

func NewChain(filters ...Filter) *Chain {
	return &Chain{filters: filters}
}

func (c *Chain) Execute(ctx *core.Context, handler Handler) error {
	for _, f := range c.filters {
		if err := f.Before(ctx); err != nil {
			return err
		}
	}

	err := handler(ctx)

	for _, f := range c.filters {
		_ = f.After(ctx)
	}

	return err
}
