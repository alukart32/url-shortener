package gzipx

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testResponse = "Gzip Test Response "
)

func TestGzip(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	r := gin.New()
	gzipHandler, err := NewGzip(gzip.BestSpeed)
	require.NoError(t, err)
	r.Use(gzipHandler.Handle)
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	r.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Header().Get("Content-Encoding"), "gzip")
	assert.NotEqual(t, w.Header().Get("Content-Length"), "0")
	assert.NotEqual(t, w.Body.Len(), len(testResponse))
	assert.Equal(t, fmt.Sprint(w.Body.Len()), w.Header().Get("Content-Length"))

	gr, err := gzip.NewReader(w.Body)
	assert.NoError(t, err)
	defer gr.Close()

	body, _ := io.ReadAll(gr)
	assert.Equal(t, string(body), testResponse)
}

func TestNoGzip(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)

	w := httptest.NewRecorder()
	r := gin.New()
	gzipHandler, err := NewGzip(gzip.BestSpeed)
	require.NoError(t, err)
	r.Use(gzipHandler.Handle)
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Length", strconv.Itoa(len(testResponse)))
		c.String(200, testResponse)
	})
	r.ServeHTTP(w, req)

	assert.Equal(t, w.Code, 200)
	assert.Equal(t, w.Header().Get("Content-Encoding"), "")
	assert.Equal(t, w.Header().Get("Content-Length"), fmt.Sprint(len(testResponse)))
	assert.Equal(t, w.Body.String(), testResponse)
}

func TestDecompressGzip(t *testing.T) {
	buf := &bytes.Buffer{}
	gz, _ := gzip.NewWriterLevel(buf, gzip.DefaultCompression)
	if _, err := gz.Write([]byte(testResponse)); err != nil {
		gz.Close()
		t.Fatal(err)
	}
	gz.Close()

	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/", buf)
	req.Header.Add("Content-Encoding", "gzip")

	r := gin.New()
	gzipHandler, err := NewGzip(gzip.BestSpeed)
	require.NoError(t, err)
	r.Use(gzipHandler.Handle)
	r.POST("/", func(c *gin.Context) {
		if v := c.Request.Header.Get("Content-Encoding"); v != "" {
			t.Errorf("unexpected `Content-Encoding`: %s header", v)
		}
		if v := c.Request.Header.Get("Content-Length"); v != "" {
			t.Errorf("unexpected `Content-Length`: %s header", v)
		}
		data, err := c.GetRawData()
		if err != nil {
			t.Fatal(err)
		}
		c.Data(200, "text/plain", data)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Header().Get("Content-Encoding"))
	assert.Equal(t, testResponse, w.Body.String())
	assert.Equal(t, "", w.Header().Get("Content-Length"))
}

func TestDecompressGzipWithEmptyBody(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/", nil)
	req.Header.Add("Content-Encoding", "gzip")

	r := gin.New()
	gzipHandler, err := NewGzip(gzip.BestSpeed)
	require.NoError(t, err)
	r.Use(gzipHandler.Handle)
	r.POST("/", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Header().Get("Content-Encoding"))
	assert.Equal(t, "ok", w.Body.String())
	assert.Equal(t, "", w.Header().Get("Content-Length"))
}

func TestDecompressGzipWithIncorrectData(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/", bytes.NewReader([]byte(testResponse)))
	req.Header.Add("Content-Encoding", "gzip")

	r := gin.New()
	gzipHandler, err := NewGzip(gzip.BestSpeed)
	require.NoError(t, err)
	r.Use(gzipHandler.Handle)
	r.POST("/", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
