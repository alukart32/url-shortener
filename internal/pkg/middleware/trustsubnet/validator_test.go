package trustsubnet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_NoTrustedSubnet(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/internal/stats", nil)
	req.Header.Add("X-Real-IP", "192.168.1.1")

	w := httptest.NewRecorder()
	r := gin.New()

	handler, err := Validator("")
	require.NoError(t, err)

	r.GET("/api/internal/stats", handler.Handle(func(c *gin.Context) {
		c.String(http.StatusOK, "response")
	}))
	r.ServeHTTP(w, req)

	assert.EqualValues(t, http.StatusForbidden, w.Code)
}

func TestValidator_TrustedSubnet(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/internal/stats", nil)
	req.Header.Add("X-Real-IP", "192.168.1.1")

	w := httptest.NewRecorder()
	r := gin.New()

	handler, err := Validator("192.168.1.1/24")
	require.NoError(t, err)

	r.GET("/api/internal/stats", handler.Handle(func(c *gin.Context) {
		c.String(http.StatusOK, "response")
	}))
	r.ServeHTTP(w, req)

	assert.EqualValues(t, http.StatusOK, w.Code)
}

func TestValidator_UntrustedSubnets(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/internal/stats", nil)
	req.Header.Add("X-Real-IP", "192.168.2.1")

	w := httptest.NewRecorder()
	r := gin.New()

	handler, err := Validator("192.168.1.1/24")
	require.NoError(t, err)

	r.GET("/api/internal/stats", handler.Handle(func(c *gin.Context) {
		c.String(http.StatusOK, "response")
	}))
	r.ServeHTTP(w, req)

	assert.EqualValues(t, http.StatusForbidden, w.Code)
}

func TestValidator_NoTrustedSubnets(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/internal/stats", nil)

	w := httptest.NewRecorder()
	r := gin.New()

	handler, err := Validator("192.168.1.1/24")
	require.NoError(t, err)

	r.GET("/api/internal/stats", handler.Handle(func(c *gin.Context) {
		c.String(http.StatusOK, "response")
	}))
	r.ServeHTTP(w, req)

	assert.EqualValues(t, http.StatusForbidden, w.Code)
}

func TestValidator_InvalidSubnet(t *testing.T) {
	_, err := Validator("192.168.1.1/76")
	require.Error(t, err)
}
