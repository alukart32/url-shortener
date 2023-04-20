// Package middleware provides gin middleware handlers.
package middleware

import (
	"github.com/alukart32/shortener-url/internal/pkg/middleware/gzipx"
	"github.com/alukart32/shortener-url/internal/pkg/middleware/logger"
	"github.com/gin-gonic/gin"
)

// Gzip returns middleware to enable GZIP support.
func Gzip(level int) (gin.HandlerFunc, error) {
	gz, err := gzipx.NewGzip(level)
	return gz.Handle, err
}

// Zerolog reutns middleware to enable custom gin router logging.
func Zerolog() gin.HandlerFunc {
	h := logger.New()
	return h.Handle
}
