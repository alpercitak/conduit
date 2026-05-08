package core

import (
	"net/http"
	"time"
)

type State map[string]any

type TraceEvent struct {
	Name      string
	Timestamp time.Time
	Duration  time.Duration
}

type Context struct {
	RequestID string
	StartTime time.Time
	Request   *http.Request

	StatusCode int

	UpstreamID  string
	UpstreamURL string

	State State
	Trace []TraceEvent
}

func (c *Context) Mark(event string) {
	now := time.Now()

	var duration time.Duration
	if len(c.Trace) > 0 {
		last := c.Trace[len(c.Trace)-1]
		duration = now.Sub(last.Timestamp)
	}

	c.Trace = append(c.Trace, TraceEvent{
		Name:      event,
		Timestamp: now,
		Duration:  duration,
	})
}
