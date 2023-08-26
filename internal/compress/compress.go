package compress

import (
	"net/http"
	"strings"
)

// CompressedTypes is types that support compression.
const compressedTypes = "application/json,text/html"

const gzipEncoding = "gzip"
const contentEncoding = "Content-Encoding"

// CompressMiddleware - the middleware compresses outgoing requests,
// if compression is supported by the client, also decompresses incoming requests.
func CompressMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(compressedTypes, contentType) {
			h.ServeHTTP(w, r)
		}

		origWriter := w

		acceptEncodings := r.Header.Values("Accept-Encoding")
		for _, acceptEncoding := range acceptEncodings {
			if strings.Contains(acceptEncoding, gzipEncoding) {
				gzipWriter := NewGzipWriter(w)
				origWriter = gzipWriter
				defer gzipWriter.Close()
			}
		}

		contentEncodings := r.Header.Values(contentEncoding)

		for _, contentEncoding := range contentEncodings {
			if strings.Contains(contentEncoding, gzipEncoding) {
				compressReader, err := NewGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = compressReader
				defer compressReader.Close()
			}
		}

		h.ServeHTTP(origWriter, r)
	})
}
