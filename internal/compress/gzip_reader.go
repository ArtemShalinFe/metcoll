package compress

import (
	"compress/gzip"
	"fmt"
	"io"
)

// gzipReader the type is used to read compressed queries.
type gzipReader struct {
	r    io.ReadCloser
	zipR *gzip.Reader
}

// NewGzipReader - Object Constructor.
func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zipR, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip new reader err: %w", err)
	}

	return &gzipReader{
		r:    r,
		zipR: zipR,
	}, nil
}

func (c gzipReader) Read(p []byte) (int, error) {
	n, err := c.zipR.Read(p)
	if err != nil {
		return 0, fmt.Errorf("an error occured while zipR reading, err: %w", err)
	}
	return n, nil
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("gzip reader parrent reader close err: %w", err)
	}
	if err := c.zipR.Close(); err != nil {
		return fmt.Errorf("gzip reader close err: %w", err)
	}

	return nil
}
