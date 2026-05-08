package balancer

import "github.com/alpercitak/conduit/apps/proxy/internal/upstream"

type Balancer interface {
	Pick(ups []*upstream.Upstream) *upstream.Upstream
}
