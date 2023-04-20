// Package gzipx provides gin middleware to enable GZIP support.
package gzipx

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// gzipWriter represents http.ResponseWriter that gzip data.
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// WriteString writes a string to writer.
func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write([]byte(s))
}

// Write writes a bytes to writer.
func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

// WriteHeader writes a header to response.
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

// NewGzip returns a new GzipHandler.
func NewGzip(level int) (*GzipHandler, error) {
	gz, err := gzip.NewWriterLevel(io.Discard, level)
	if err != nil {
		return nil, err
	}

	return &GzipHandler{
		writer: gz,
	}, nil
}

// GzipHandler represents a gin GZIP handler.
type GzipHandler struct {
	writer *gzip.Writer
}

// Handle handles incoming request.
func (h *GzipHandler) Handle(c *gin.Context) {
	if c.Request.Header.Get("Content-Encoding") == "gzip" {
		decompress(c)
	}

	if !shouldCompress(c.Request, "gzip") {
		c.Next()
		return
	}

	h.writer.Reset(c.Writer)
	c.Header("Content-Encoding", "gzip")
	c.Writer = &gzipWriter{c.Writer, h.writer}
	defer func() {
		h.writer.Close()
		c.Header("Content-Length", fmt.Sprint(c.Writer.Size()))
	}()
	c.Next()
}

// shouldCompress checks the ability to compress for incoming request.
func shouldCompress(req *http.Request, format string) bool {
	if strings.Contains(req.Header.Get("Accept-Encoding"), format) &&
		isCompressibleMIME(req.Header.Get("Content-Type")) {
		return true
	}
	return false
}

// decompress reades a request body and decompress it.
func decompress(c *gin.Context) {
	if c.Request.Body == nil {
		return
	}
	r, err := gzip.NewReader(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.Request.Header.Del("Content-Encoding")
	c.Request.Header.Del("Content-Length")
	c.Request.Body = r
}
