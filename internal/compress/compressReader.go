package compress

import (
	"compress/gzip"
	"io"
)

type CompressReader struct {
	r    io.ReadCloser
	zipR *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zipR, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:    r,
		zipR: zipR,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zipR.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zipR.Close()
}
