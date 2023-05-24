package compress

import (
	"compress/gzip"
	"net/http"
)

type CompressWriter struct {
	http.ResponseWriter
	zipW *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		zipW:           gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.zipW.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.zipW.Close()
}
