package core

import (
	"net/http"
	"time"
)

type State map[string]any

type TraceEvent struct {
	Name      string
	Timestamp time.Time
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
	c.Trace = append(c.Trace, TraceEvent{
		Name:      event,
		Timestamp: time.Now(),
	})
}
