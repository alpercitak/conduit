package core

import (
	"net/http"
	"time"
)

type Context struct {
	RequestID string
	StartTime time.Time
	Request   *http.Request
}
