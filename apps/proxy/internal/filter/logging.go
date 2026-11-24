package filter

import (
	"log"
	"time"

	"proxy/internal/core"
)

type LoggingFilter struct{}

func NewLoggingFilter() *LoggingFilter {
	return &LoggingFilter{}
}

func (f *LoggingFilter) Before(ctx *core.Context) error {
	return nil
}

// After logs one line per request; the ID is the key for /trace?id=.
func (f *LoggingFilter) After(ctx *core.Context) error {
	log.Printf("%s %s %s upstream=%s status=%d took=%s",
		ctx.RequestID, ctx.Request.Method, ctx.Request.URL.Path,
		ctx.UpstreamID, ctx.StatusCode, time.Since(ctx.StartTime))
	return nil
}
