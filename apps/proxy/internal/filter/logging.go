package filter

import (
	"log"
	"time"

	"github.com/alpercitak/conduit/apps/proxy/internal/core"
)

type LoggingFilter struct{}

func NewLoggingFilter() *LoggingFilter {
	return &LoggingFilter{}
}

func (f *LoggingFilter) Before(ctx *core.Context) error {
	log.Printf("[start] %s %s", ctx.RequestID, ctx.Request.URL.Path)
	return nil
}

func (f *LoggingFilter) After(ctx *core.Context) error {
	duration := time.Since(ctx.StartTime)
	log.Printf("[end] %s took=%s", ctx.RequestID, duration)
	return nil
}
