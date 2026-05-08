package filter

import "github.com/alpercitak/conduit/apps/proxy/internal/core"

type Chain struct {
	filters []Filter
}

func NewChain(filters ...Filter) *Chain {
	return &Chain{filters: filters}
}

func (c *Chain) Before(ctx *core.Context) error {
	for _, f := range c.filters {
		if err := f.Before(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) After(ctx *core.Context) error {
	for _, f := range c.filters {
		if err := f.After(ctx); err != nil {
			return err
		}
	}
	return nil
}
