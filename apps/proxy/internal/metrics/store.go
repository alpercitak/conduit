package metrics

import (
	"sort"
	"sync"
	"time"
)

type Store struct {
	mu sync.Mutex

	Requests int
	Errors   int

	latencies []time.Duration
}

func NewStore() *Store {
	return &Store{
		latencies: make([]time.Duration, 0, 1024),
	}
}

func (s *Store) Record(success bool, latency time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Requests++

	if !success {
		s.Errors++
	}

	s.latencies = append(s.latencies, latency)
}

func (s *Store) snapshot() []time.Duration {
	cp := make([]time.Duration, len(s.latencies))
	copy(cp, s.latencies)
	return cp
}

func percentile(data []time.Duration, p float64) time.Duration {
	if len(data) == 0 {
		return 0
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i] < data[j]
	})

	index := int(float64(len(data)) * p)
	if index >= len(data) {
		index = len(data) - 1
	}

	return data[index]
}

func (s *Store) Summary() (p50, p95, p99 time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := s.snapshot()

	if len(data) == 0 {
		return 0, 0, 0
	}

	p50 = percentile(data, 0.50)
	p95 = percentile(data, 0.95)
	p99 = percentile(data, 0.99)

	return
}
