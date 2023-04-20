// Package ginx provides an extended gin.Engine.
package ginx

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/alukart32/shortener-url/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

var (
	once sync.Once
	g    *gin.Engine
)

// Get returns extended gin.Engine instance.
func Get() (*gin.Engine, error) {
	var err error

	once.Do(func() {
		mode := os.Getenv("GIN_MODE")
		if len(mode) == 0 {
			mode = gin.DebugMode // default to debug mode
		}

		gin.SetMode(mode)

		g = gin.New()

		// gin Logger: os.Stdout (default)
		g.Use(middleware.Zerolog())

		// gzip middleware
		gzipHandler, err := middleware.Gzip(gzip.BestSpeed)
		if err != nil {
			return
		}
		g.Use(gzipHandler)

		// gin Recovery
		g.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
			if err, ok := recovered.(string); ok {
				c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
			}
			c.AbortWithStatus(http.StatusInternalServerError)
		}))
	})

	return g, err
}
