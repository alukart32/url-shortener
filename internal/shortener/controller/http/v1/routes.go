// Package v1 provides routes for HTTP API v1 routes.
package v1

import (
	"io"
	"net/http"

	"github.com/alukart32/shortener-url/internal/shortener/services"
	"github.com/gin-gonic/gin"
)

// userHandler defines the gin.HandlerFunc with the userID param.
type userHandler func(c *gin.Context, userID string)

// authHandler defines the HTTP auth handler for userHandler.
type authHandler interface {
	Handle(func(c *gin.Context, userID string)) gin.HandlerFunc
}

// validateSubnetHandler defines the validator of the remote subnet request address.
type validateSubnetHandler interface {
	Handle(next gin.HandlerFunc) gin.HandlerFunc
}

// SetRoutes adds HTTP API v1 routes to the gin router.
func SetRoutes(
	g *gin.Engine,
	auth authHandler,
	subnetValidator validateSubnetHandler,
	servs *services.Services,
) {
	// Add short URLs handler.
	g.POST("/", auth.Handle((short(servs.Shortener, servs.Provider))))
	g.POST("/api/shorten", auth.Handle(shorten(servs.Shortener, servs.Provider)))

	// Add get by slug handler.
	g.GET("/:slug", getBySlug(servs.Provider))

	// Add collect URLs handler.
	g.GET("/api/user/urls", auth.Handle(collectURLs(servs.Provider)))

	// Add delete URLs handler.
	g.DELETE("/api/user/urls", auth.Handle(deleteURLs(servs.Deleter)))

	// Add batch URLs handler.
	g.POST("/api/shorten/batch", auth.Handle(batchURLs(servs.Shortener)))

	// Add stat handler.
	g.GET("/api/internal/stats", subnetValidator.Handle((stat(servs.Statistic))))
}

// readReqBody reads the request body.
func readReqBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
