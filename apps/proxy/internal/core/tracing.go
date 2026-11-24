package core

import (
	"sync"
	"time"
)

type TraceEvent struct {
	Name      string
	Timestamp time.Time
	Duration  time.Duration
}

type Tracer interface {
	Mark(ctx *Context, name string)
	Track(ctx *Context, name string, fn func())
}

// maxTraces bounds memory: only the most recent request IDs are kept.
const maxTraces = 10000

type InMemoryTracer struct {
	traces map[string][]TraceEvent
	mu     sync.Mutex

	limit int
	order []string // ring of IDs in insertion order
	next  int      // oldest slot once the ring is full
}

func NewInMemoryTracer() *InMemoryTracer {
	return &InMemoryTracer{
		traces: make(map[string][]TraceEvent),
		limit:  maxTraces,
	}
}

func (t *InMemoryTracer) append(id string, evt TraceEvent) {
	if _, ok := t.traces[id]; !ok {
		if len(t.order) < t.limit {
			t.order = append(t.order, id)
		} else {
			delete(t.traces, t.order[t.next])
			t.order[t.next] = id
			t.next = (t.next + 1) % t.limit
		}
	}
	t.traces[id] = append(t.traces[id], evt)
}

func (t *InMemoryTracer) Mark(ctx *Context, name string) {
	t.mu.Lock()
	t.append(ctx.RequestID, TraceEvent{
		Name:      name,
		Timestamp: time.Now(),
	})
	t.mu.Unlock()
}

func (t *InMemoryTracer) Track(ctx *Context, name string, fn func()) {
	start := time.Now()
	fn()

	t.mu.Lock()
	t.append(ctx.RequestID, TraceEvent{
		Name:      name,
		Timestamp: start,
		Duration:  time.Since(start),
	})
	t.mu.Unlock()
}

func (t *InMemoryTracer) Get(id string) ([]TraceEvent, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	trace, ok := t.traces[id]
	return trace, ok
}
