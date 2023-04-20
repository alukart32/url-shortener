package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// deleter defines the shortened URLs deleter by slugs.
type deleter interface {
	Delete(userID string, slugs []string) error
}

// deleteURLs returns a new handler for the delete URLs route.
func deleteURLs(deleter deleter) userHandler {
	return func(c *gin.Context, userID string) {
		body, err := readReqBody(c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		var slugs []string
		if err = json.Unmarshal(body, &slugs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		if len(slugs) == 0 {
			c.Status(http.StatusNoContent)
			return
		}

		if err = deleter.Delete(userID, slugs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.Writer.WriteHeader(http.StatusAccepted)
	}
}
