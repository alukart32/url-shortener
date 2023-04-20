package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/gin-gonic/gin"
)

// collector defines the shortened URLs Collector.
type collector interface {
	CollectByUser(context.Context, string) ([]models.ShortenedURL, error)
}

// collectURLsResponse defines the response to the CollectURLs request.
type collectURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// collectURLs returns a new handler for the collect URLs route.
func collectURLs(collector collector) userHandler {
	return func(c *gin.Context, userID string) {
		urls, err := collector.CollectByUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if len(urls) == 0 {
			c.Status(http.StatusNoContent)
			return
		}

		list := make([]collectURLsResponse, len(urls))
		for i, u := range urls {
			list[i] = collectURLsResponse{
				ShortURL:    u.Value,
				OriginalURL: u.Raw,
			}
		}

		respBody, err := json.Marshal(list)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", respBody)
	}
}
