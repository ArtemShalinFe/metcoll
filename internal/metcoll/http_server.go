package metcoll

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/crypto"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"go.uber.org/zap"
)

type HTTPServer struct {
	httpServer    *http.Server
	log           *zap.SugaredLogger
	trustedSubnet *net.IPNet
	privateKey    []byte
	hashkey       []byte
}

// NewHTTPServer - Object Constructor.
func NewHTTPServer(
	ctx context.Context,
	stg Storage,
	cfg *configuration.Config,
	sl *zap.SugaredLogger,
) (*HTTPServer, error) {
	var privateKey []byte
	if cfg.PrivateCryptoKey != "" {
		privateCryptoKey, err := crypto.GetKeyBytes(cfg.PrivateCryptoKey)
		if err != nil {
			return nil, fmt.Errorf("an occured error when server getting key bytes, err: %w", err)
		}
		privateKey = privateCryptoKey
	}

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		return nil, fmt.Errorf("cannot init middleware logger err: %w ", err)
	}

	s := http.Server{
		Addr: cfg.Address,
	}

	srv := &HTTPServer{
		&s,
		sl,
		parseTrustedSubnet(cfg.TrustedSubnet),
		privateKey,
		cfg.Key,
	}

	srv.httpServer.Handler = NewRouter(ctx,
		NewHandler(stg, sl),
		srv.resolverIP,
		l.RequestLogger,
		srv.requestHashChecker,
		srv.responseHashSetter,
		compress.CompressMiddleware,
		srv.cryptoDecrypter)

	return srv, nil
}

func (s *HTTPServer) ListenAndServe() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server listen and serve err: %w", err)
		}
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown err: %w", err)
	}
	return nil
}

// IPResolver - middleware checks header X-Real-IP in the incoming request.
func (s *HTTPServer) resolverIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.trustedSubnet == nil {
			h.ServeHTTP(w, r)
			return
		}

		ipStr := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if strings.TrimSpace(ipStr) == "" {
			w.WriteHeader(http.StatusBadRequest)
			s.log.Info("header X-Real-IP is empty")
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			w.WriteHeader(http.StatusBadRequest)
			s.log.Info("failed parse ip from http header")
			return
		}

		if !s.trustedSubnet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			s.log.Info("trusted network does not contain the ip address value from the 'X-Real-IP' header")
			return
		}
	})
}

// CryptoDecrypter - middleware decrypt the incoming request.
func (s *HTTPServer) cryptoDecrypter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(s.privateKey) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		var cryptBuf bytes.Buffer
		tee := io.TeeReader(r.Body, &cryptBuf)
		body, err := io.ReadAll(tee)
		if err != nil {
			s.log.Errorf("an occured error when reading body err: %w", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		decrypted, err := crypto.Decrypt(s.privateKey, body)
		if err != nil {
			s.log.Errorf("an occured error when decrypt body err: %w", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var decryptBuf bytes.Buffer
		_, err = decryptBuf.Write(decrypted)
		if err != nil {
			s.log.Errorf("an occured error when write decrypted body err: %w", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(&decryptBuf)
	})
}

// RequestHashChecker - middleware checks the hash in the incoming request.
func (s *HTTPServer) requestHashChecker(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(s.hashkey) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		// for ya-autotests.
		bodyHash := r.Header.Get(HashSHA256)
		if bodyHash == "" {
			h.ServeHTTP(w, r)
			return
		}

		var buf bytes.Buffer
		tee := io.TeeReader(r.Body, &buf)
		body, err := io.ReadAll(tee)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(&buf)

		hash := hmac.New(sha256.New, s.hashkey)
		hash.Write(body)

		sign := hash.Sum(nil)

		if hashBytesToString(hash, sign) == bodyHash {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, "incorrect hash", http.StatusBadRequest)
			return
		}
	})
}

// ResponceHashSetter - middleware sets the hash in the server response.
func (s *HTTPServer) responseHashSetter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(s.hashkey) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		hsw := newResponseHashSetter(w, s.hashkey)

		h.ServeHTTP(hsw, r)
	})
}

type ResponseHashWriter struct {
	http.ResponseWriter
	hashkey []byte
}

// NewResponseHashSetter - Object Constructor.
func newResponseHashSetter(w http.ResponseWriter, hashkey []byte) *ResponseHashWriter {
	return &ResponseHashWriter{
		ResponseWriter: w,
		hashkey:        hashkey,
	}
}

func (r *ResponseHashWriter) Write(b []byte) (int, error) {
	hash := hmac.New(sha256.New, r.hashkey)
	hash.Write(b)

	r.ResponseWriter.Header().Set(HashSHA256, hashBytesToString(hash, nil))

	n, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("response write was faild, err: %w", err)
	}

	return n, nil
}

func parseTrustedSubnet(trustedSubnet string) *net.IPNet {
	if trustedSubnet == "" {
		return nil
	}

	ip := net.ParseIP(trustedSubnet)
	if ip == nil {
		return nil
	}
	mask := ip.DefaultMask()
	if mask == nil {
		return nil
	}
	tn := &net.IPNet{
		IP:   ip,
		Mask: mask,
	}

	return tn
}
