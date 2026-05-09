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
	Flush(ctx *Context) []TraceEvent
}

type InMemoryTracer struct {
	traces map[string][]TraceEvent
	mu     sync.Mutex
}

func NewInMemoryTracer() *InMemoryTracer {
	return &InMemoryTracer{
		traces: make(map[string][]TraceEvent),
	}
}

func (t *InMemoryTracer) append(id string, evt TraceEvent) {
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

func (t *InMemoryTracer) Flush(ctx *Context) []TraceEvent {
	t.mu.Lock()
	defer t.mu.Unlock()

	events, ok := t.traces[ctx.RequestID]
	if !ok {
		return nil
	}

	delete(t.traces, ctx.RequestID)
	return events
}
