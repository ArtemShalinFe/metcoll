package compress

import (
	"compress/gzip"
	"fmt"
	"net/http"
)

// gzipReader the type is used to write compressed queries.
type gzipWriter struct {
	http.ResponseWriter
	zipW *gzip.Writer
}

// NewGzipWriter - Object Constructor.
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		zipW:           gzip.NewWriter(w),
	}
}

func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.zipW.Write(p)
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.ResponseWriter.Header().Set(contentEncoding, gzipEncoding)
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	if err := c.zipW.Close(); err != nil {
		return fmt.Errorf("gzip writer close err: %w", err)
	}

	return nil
}
