package core

import (
	"net/http"
	"time"
)

type State map[string]any

type Context struct {
	RequestID string
	StartTime time.Time
	Request   *http.Request

	StatusCode int

	UpstreamID  string
	UpstreamURL string

	State State

	Outcome Outcome
}
