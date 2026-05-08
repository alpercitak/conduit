package upstream

import "github.com/alpercitak/conduit/apps/proxy/internal/circuit"

type Upstream struct {
	ID      string
	URL     string
	Breaker *circuit.Breaker
}
