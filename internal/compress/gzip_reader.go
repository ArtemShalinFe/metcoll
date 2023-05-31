package compress

import (
	"compress/gzip"
	"io"
)

type gzipReader struct {
	r    io.ReadCloser
	zipR *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zipR, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:    r,
		zipR: zipR,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zipR.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zipR.Close()
}
