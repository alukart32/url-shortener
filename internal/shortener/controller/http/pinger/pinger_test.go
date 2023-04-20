package pinger

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pingerMock struct {
	PingFn func() error
}

func (m *pingerMock) Ping() error {
	if m != nil && m.PingFn != nil {
		return m.PingFn()
	}
	return fmt.Errorf("unable to ping")
}

func TestPinger_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := gin.New()
	PostgresPinger(g, &pingerMock{PingFn: func() error { return nil }})

	w := httptest.NewRecorder()
	// Prepare request.
	req := textReq(t, "/ping", http.MethodGet, nil)
	g.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	assert.EqualValues(t, http.StatusOK, resp.StatusCode,
		"Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
}

func TestPinger_InternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := gin.New()
	PostgresPinger(g, &pingerMock{})

	w := httptest.NewRecorder()
	// Prepare request.
	req := textReq(t, "/ping", http.MethodGet, nil)
	g.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	assert.EqualValues(t, http.StatusInternalServerError, resp.StatusCode,
		"Expected status code: %d, got %d", http.StatusOK, resp.StatusCode)
}

// textReq prepares http.Request with a plain text body.
func textReq(t *testing.T, api, method string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, api, body)
	require.NoError(t, err)

	return req
}
