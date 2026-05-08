package upstream

import "errors"

type Pool struct {
	upstreams []*Upstream
	index     map[string]*Upstream
}

func NewPool(ups []*Upstream) *Pool {
	m := make(map[string]*Upstream)

	for _, u := range ups {
		m[u.ID] = u
	}

	return &Pool{
		upstreams: ups,
		index:     m,
	}
}

func (p *Pool) GetByID(id string) *Upstream {
	return p.index[id]
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
