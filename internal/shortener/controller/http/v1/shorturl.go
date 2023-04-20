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

// shortener defines the url Shortener.
type shortener interface {
	Short(context.Context, models.URL) (string, error)
}

// getterByURL defines the shortened URL provider by raw url.
type getterByURL interface {
	GetByURL(context.Context, string) (models.ShortenedURL, error)
}

// short returns a new handler for the short URL route.
func short(shortener shortener, provider getterByURL) userHandler {
	return func(c *gin.Context, userID string) {
		reqData, err := readReqBody(c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		shortenedURL, err := shortener.Short(c.Request.Context(),
			models.NewURL(userID, "", string(reqData)))
		if err != nil {
			if errors.Is(err, shorturl.ErrUniqueViolation) {
				// Get an existing shortened URL.
				shortURL, err := provider.GetByURL(c.Request.Context(), string(reqData))
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
				}
				c.String(http.StatusConflict, shortURL.Value)
				return
			}
			respStatus := http.StatusInternalServerError
			if errors.Is(err, shorturl.ErrInvalidCreation) {
				respStatus = http.StatusBadRequest
			}
			c.JSON(respStatus, gin.H{"error": err})
			return
		}

		c.String(http.StatusCreated, shortenedURL)
	}
}

// shortURLRequest is a request to short the URL.
type shortURLRequest struct {
	URL string `json:"url"`
}

// shortURLResponse is a response to short the URL.
type shortURLResponse struct {
	Result string `json:"result"`
}

// shorten returns a new handler for the short URL route.
func shorten(shortener shortener, provider getterByURL) userHandler {
	return func(c *gin.Context, userID string) {
		reqBody, err := readReqBody(c.Request)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		var reqData shortURLRequest
		if err := json.Unmarshal(reqBody, &reqData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		shortenURL, err := shortener.Short(c.Request.Context(),
			models.NewURL(userID, "", reqData.URL))
		if err != nil {
			if errors.Is(err, shorturl.ErrUniqueViolation) {
				// Get an existing shortened URL.
				shortenedURL, err := provider.GetByURL(c.Request.Context(), reqData.URL)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
				}
				respBody, err := json.Marshal(shortURLResponse{
					Result: shortenedURL.Value,
				})
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				c.Data(http.StatusConflict, "application/json; charset=utf-8", respBody)
				return
			}
			respStatus := http.StatusInternalServerError
			if errors.Is(err, shorturl.ErrInvalidCreation) {
				respStatus = http.StatusBadRequest
			}
			c.JSON(respStatus, gin.H{"error": err})
			return
		}
		respBody, err := json.Marshal(shortURLResponse{
			Result: shortenURL,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.Data(http.StatusCreated, "application/json; charset=utf-8", respBody)
	}
}
