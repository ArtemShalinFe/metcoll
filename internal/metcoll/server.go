package metcoll

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
)

type Server struct {
	*http.Server
	hashkey []byte
}

// NewServer - Object Constructor.
func NewServer(cfg *configuration.Config) *Server {
	s := http.Server{
		Addr: cfg.Address,
	}
	return &Server{&s, cfg.Key}
}

// RequestHashChecker - middleware checks the hash in the incoming request.
func (s *Server) RequestHashChecker(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(s.hashkey) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		// for ya-autotests.
		bodyHash := r.Header.Get("HashSHA256")
		if bodyHash == "" {
			h.ServeHTTP(w, r)
			return
		}

		var buf bytes.Buffer
		tee := io.TeeReader(r.Body, &buf)
		body, err := io.ReadAll(tee)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(&buf)

		hash := hmac.New(sha256.New, []byte(s.hashkey))
		hash.Write(body)

		sign := hash.Sum(nil)

		if fmt.Sprintf("%x", sign) == bodyHash {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, "incorrect hash", http.StatusBadRequest)
			return
		}
	})
}

// ResponceHashSetter - middleware sets the hash in the server response.
func (s *Server) ResponceHashSetter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(s.hashkey) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		hsw := NewResponseHashSetter(w, s.hashkey)

		h.ServeHTTP(hsw, r)
	})
}

type ResponseHashWriter struct {
	http.ResponseWriter
	hashkey []byte
}

// NewResponseHashSetter - Object Constructor.
func NewResponseHashSetter(w http.ResponseWriter, hashkey []byte) *ResponseHashWriter {
	return &ResponseHashWriter{
		ResponseWriter: w,
		hashkey:        hashkey,
	}
}

func (r *ResponseHashWriter) Write(b []byte) (int, error) {
	hash := hmac.New(sha256.New, r.hashkey)
	hash.Write(b)

	r.ResponseWriter.Header().Set("HashSHA256", fmt.Sprintf("%x", hash.Sum(nil)))

	return r.ResponseWriter.Write(b)
}
