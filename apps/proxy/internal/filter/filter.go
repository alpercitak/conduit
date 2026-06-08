package filter

import "proxy/internal/core"

type Filter interface {
	Before(ctx *core.Context) error
	After(ctx *core.Context) error
}
