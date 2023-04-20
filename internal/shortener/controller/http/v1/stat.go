package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/gin-gonic/gin"
)

// statProvider defines the shortened URLs statistics provider.
type statProvider interface {
	Stat(context.Context) (models.Stat, error)
}

// stat returns a new shortened URLs statistics handler.
func stat(provider statProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		stat, err := provider.Stat(c.Request.Context())
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		respData := statResponse{
			URLs:  stat.URLs,
			Users: stat.Users,
		}
		respBody, err := json.Marshal(respData)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Data(http.StatusOK, "application/json; charset=utf-8", respBody)
	}
}

type statResponse struct {
	URLs  int `json:"urls"`  // количество сокращённых URL в сервисе
	Users int `json:"users"` // количество пользователей в сервисе
}
