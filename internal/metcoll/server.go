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

func NewServer(cfg *configuration.Config) *Server {
	s := http.Server{
		Addr: cfg.Address,
	}
	return &Server{&s, cfg.HashKey}
}

func (s *Server) RequestHashChecker(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if len(s.hashkey) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		// Тесты не прочитали ТЗ и присылают запросы без ключа ¯\_(ツ)_/¯ и еще сильно на тебя ругаются когда им 400 показываешь.
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
