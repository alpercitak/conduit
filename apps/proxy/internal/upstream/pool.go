package upstream

type Pool struct {
	upstreams []*Upstream
	index     map[string]*Upstream
}

func NewPool(ups []*Upstream) *Pool {
	index := make(map[string]*Upstream)

	for _, u := range ups {
		index[u.ID] = u
	}

	return &Pool{
		upstreams: ups,
		index:     index,
	}
}

func (p *Pool) All() []*Upstream {
	return p.upstreams
}

func (p *Pool) GetByID(id string) *Upstream {
	return p.index[id]
}
