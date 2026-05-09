package metrics

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const maxBufferSize = 50_000

type Store struct {
	mu sync.Mutex

	requests int64
	errors   int64

	latencies []time.Duration
}

func NewStore() *Store {
	return &Store{
		latencies: make([]time.Duration, 0, 1024),
	}
}

func (s *Store) Record(success bool, latency time.Duration) {
	atomic.AddInt64(&s.requests, 1)

	if !success {
		atomic.AddInt64(&s.errors, 1)
	}

	s.mu.Lock()
	if len(s.latencies) < maxBufferSize {
		s.latencies = append(s.latencies, latency)
	}
	s.mu.Unlock()
}

func calcPercentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

// Summary returns and clears the latency buffer
func (s *Store) Summary() (p50, p95, p99 time.Duration, reqs, errs int64) {
	s.mu.Lock()

	if len(s.latencies) == 0 {
		s.mu.Unlock()
		return 0, 0, 0,
			atomic.LoadInt64(&s.requests),
			atomic.LoadInt64(&s.errors)
	}

	// snapshot + reset buffer
	data := append([]time.Duration(nil), s.latencies...)
	s.latencies = make([]time.Duration, 0, 1024)

	s.mu.Unlock()

	reqs = atomic.LoadInt64(&s.requests)
	errs = atomic.LoadInt64(&s.errors)

	slices.Sort(data)

	return calcPercentile(data, 0.50),
		calcPercentile(data, 0.95),
		calcPercentile(data, 0.99),
		reqs,
		errs
}
