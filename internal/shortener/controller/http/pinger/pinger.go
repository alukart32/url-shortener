// Package pinger provides ping routes for for API version 1.
package pinger

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Pinger defines the one who can ping.
type pinger interface {
	Ping() error
}

// PostgresPinger sets the postgres route.
func PostgresPinger(h *gin.Engine, pinger pinger) error {
	if pinger != nil {
		h.GET("/ping", func(c *gin.Context) {
			if err := pinger.Ping(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			}
			c.Status(http.StatusOK)
		})
	}
	return nil
}
