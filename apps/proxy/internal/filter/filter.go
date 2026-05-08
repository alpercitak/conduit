package filter

import "github.com/alpercitak/conduit/apps/proxy/internal/core"

type Filter interface {
	Before(ctx *core.Context) error
	After(ctx *core.Context) error
}
