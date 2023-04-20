package v1

import (
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var authWrap = func(next func(c *gin.Context, userID string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.NewUUID()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		next(c, id.String())
	}
}

// textReq prepares http.Request with a plain text body.
func textReq(t *testing.T, api, method string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, api, body)
	require.NoError(t, err)

	return req
}

// setupGin sets up a new gin.Engine.
func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}
