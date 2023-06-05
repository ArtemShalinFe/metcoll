package compress

import (
	"compress/gzip"
	"net/http"
)

type gzipWriter struct {
	http.ResponseWriter
	zipW *gzip.Writer
}

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
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	return c.zipW.Close()
}
