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
	Flush(ctx *Context)
}

type InMemoryTracer struct {
	traces map[string][]TraceEvent
	mu     sync.RWMutex
}

func NewInMemoryTracer() *InMemoryTracer {
	return &InMemoryTracer{
		traces: make(map[string][]TraceEvent),
	}
}

func (t *InMemoryTracer) Mark(ctx *Context, name string) {
	evt := TraceEvent{
		Name:      name,
		Timestamp: time.Now(),
	}

	t.mu.Lock()
	t.traces[ctx.RequestID] = append(t.traces[ctx.RequestID], evt)
	t.mu.Unlock()
}

func (t *InMemoryTracer) Track(ctx *Context, name string, fn func()) {
	start := time.Now()
	fn()

	t.mu.Lock()
	t.traces[ctx.RequestID] = append(t.traces[ctx.RequestID], TraceEvent{
		Name:      name,
		Timestamp: start,
		Duration:  time.Since(start),
	})
	t.mu.Unlock()
}

func (t *InMemoryTracer) Get(id string) ([]TraceEvent, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	trace, ok := t.traces[id]
	return trace, ok
}

func (t *InMemoryTracer) Flush(ctx *Context) {

}
