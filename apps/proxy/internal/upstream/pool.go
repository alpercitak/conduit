package upstream

import "errors"

type Pool struct {
	upstreams []*Upstream
}

func NewPool(ups []*Upstream) *Pool {
	return &Pool{upstreams: ups}
}

func (p *Pool) All() []*Upstream {
	return p.upstreams
}

func (p *Pool) First() (*Upstream, error) {
	if len(p.upstreams) == 0 {
		return nil, errors.New("no upstreams available")
	}
	return p.upstreams[0], nil
}
