// Package logger provides a custom logger for the gin router.
package logger

import (
	"context"
	"time"

	"github.com/alukart32/shortener-url/internal/pkg/zerologx"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// ZerologHandler represents a logger for gin router.
type ZerologHandler struct{}

// New returns a new ZerologHandler.
func New() *ZerologHandler {
	return &ZerologHandler{}
}

// CorrID represents the corresponding ID for the request.
type CorrID string

// Handle adds zerlog context to the request context.
func (h *ZerologHandler) Handle(c *gin.Context) {
	t := time.Now()

	l := zerologx.Get()

	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery

	correlationID := xid.New().String()
	ctx := context.WithValue(c.Request.Context(), CorrID("correlation_id"), correlationID)
	c.Request = c.Request.WithContext(ctx)

	l.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("correlation_id", correlationID)
	})

	c.Request = c.Request.WithContext(l.WithContext(c.Request.Context()))

	c.Next()

	if raw != "" {
		path = path + "?" + raw
	}

	l.Info().
		Str("method", c.Request.Method).
		Str("path", path).
		Int("status", c.Writer.Status()).
		Dur("elapsed_ms", time.Since(t)).
		Msg("incoming request")
}
