package compress

import (
	"net/http"
	"strings"
)

const compressedTypes = "application/json,text/html"

func CompressMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(compressedTypes, contentType) {
			h.ServeHTTP(w, r)
		}

		origWriter := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			compressWriter := NewCompressWriter(w)
			origWriter = compressWriter
			defer compressWriter.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			compressReader, err := NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = compressReader
			defer compressReader.Close()
		}

		h.ServeHTTP(origWriter, r)

	})
}
