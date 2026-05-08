package metrics

import (
	"sync"
	"time"
)

type Store struct {
	mu sync.Mutex

	Requests int
	Errors   int

	Latencies []time.Duration
}

func NewStore() *Store {
	return &Store{
		Latencies: make([]time.Duration, 0, 1024),
	}
}

func (s *Store) Record(success bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Requests++

	if !success {
		s.Errors++
	}

	s.Latencies = append(s.Latencies, latency)
}
