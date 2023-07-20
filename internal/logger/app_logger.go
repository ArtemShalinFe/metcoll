package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"go.uber.org/zap"
)

type MiddlewareLogger struct {
	*zap.SugaredLogger
}

func NewMiddlewareLogger(l *zap.SugaredLogger) (*MiddlewareLogger, error) {

	return &MiddlewareLogger{
		l,
	}, nil

}

func (l *MiddlewareLogger) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := NewResponseLoggerWriter(w)
		var buf bytes.Buffer
		tee := io.TeeReader(r.Body, &buf)
		body, err := io.ReadAll(tee)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			l.Errorf("rlogger read body error: %w", err)
			return
		}
		r.Body = io.NopCloser(&buf)

		start := time.Now()
		h.ServeHTTP(rw, r)
		duration := time.Since(start)

		l.Infof("HTTP request method: %s, header: %v, body: %s, url: %s, duration: %s, statusCode: %d, responseSize: %d",
			r.Method, r.Header, string(body), r.RequestURI, duration, rw.responseData.status, rw.responseData.size,
		)
	})
}

func (l *MiddlewareLogger) Interrupt() error {

	if err := l.Sync(); err != nil {

		if runtime.GOOS == "darwin" {
			return nil
		} else {
			return fmt.Errorf("cannot flush buffered log entries err: %w", err)
		}

	}

	return nil

}
