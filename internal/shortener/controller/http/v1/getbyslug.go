package v1

import (
	"context"
	"net/http"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/gin-gonic/gin"
)

// getterBySlug defines the shortened URL provider by slug.
type getterBySlug interface {
	GetBySlug(context.Context, string) (models.ShortenedURL, error)
}

// New returns a new handler for the get URL by slug route.
func getBySlug(provider getterBySlug) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		if len(slug) == 0 {
			c.Status(http.StatusBadRequest)
			return
		}

		shortenedURL, err := provider.GetBySlug(c.Request.Context(), slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if shortenedURL.Empty() {
			c.Status(http.StatusNotFound)
			return
		}

		if shortenedURL.IsDeleted {
			c.Status(http.StatusGone)
			return
		}

		c.Header("Location", shortenedURL.Raw)
		c.Status(http.StatusTemporaryRedirect)
	}
}
