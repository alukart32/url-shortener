package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/services/shorturl"
	"github.com/gin-gonic/gin"
)

// batcher defines the URLs Batcher.
type batcher interface {
	Batch(context.Context, []models.URL) ([]models.ShortenedURL, error)
}

// batchURLsRequest defines the item in a request for the batch urls route.
type batchURLsRequest struct {
	CorrID string `json:"correlation_id"`
	RawURL string `json:"original_url"`
}

// batchURLsResponse defines the item in a response for the batch urls route.
type batchURLsResponse struct {
	CorrID       string `json:"correlation_id"`
	ShortenedURL string `json:"short_url"`
}

// batchURLs returns a new handler for the batch URLs route.
func batchURLs(batcher batcher) userHandler {
	return func(c *gin.Context, userID string) {
		reqBody, err := readReqBody(c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		var reqData []batchURLsRequest
		if err := json.Unmarshal(reqBody, &reqData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		urlsToBatch := make([]models.URL, len(reqData))
		for i, v := range reqData {
			urlsToBatch[i] = models.NewURL(userID, v.CorrID, v.RawURL)
		}
		urls, err := batcher.Batch(c.Request.Context(), urlsToBatch)
		if err != nil {
			if errors.Is(err, shorturl.ErrEmptyBatch) ||
				errors.Is(err, shorturl.ErrInvalidCreation) {
				c.Status(http.StatusBadRequest)
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		respData := make([]batchURLsResponse, len(urls))
		for i, u := range urls {
			respData[i] = batchURLsResponse{
				CorrID:       u.CorrID,
				ShortenedURL: u.Value,
			}
		}

		respBody, err := json.Marshal(respData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.Data(http.StatusCreated, "application/json; charset=utf-8", respBody)
	}
}
