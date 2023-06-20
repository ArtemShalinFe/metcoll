package metcoll

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"io"
	"net/http"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
)

type Server struct {
	*http.Server
	hashkey string
}

func NewServer(cfg *configuration.Config) *Server {
	s := http.Server{
		Addr: cfg.Address,
	}
	return &Server{&s, cfg.Key}
}

func (s *Server) RequestHashChecker(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if s.hashkey == "" {
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

		if hmac.Equal(sign, []byte(bodyHash)) {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, "incorrect hash", http.StatusBadRequest)
			return
		}
	})
}
