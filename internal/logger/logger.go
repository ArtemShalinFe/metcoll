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

type AppLogger struct {
	Log *zap.SugaredLogger
}

func NewLogger() (*AppLogger, error) {

	l, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("cannot init zap-logger err: %w ", err)
	}

	sl := l.Sugar()

	return &AppLogger{
		sl,
	}, nil

}

func (l *AppLogger) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := NewResponseLoggerWriter(w)

		var buf bytes.Buffer
		tee := io.TeeReader(r.Body, &buf)
		body, err := io.ReadAll(tee)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			l.Log.Errorf("rlogger read body error: %w", err)
			return
		}
		r.Body = io.NopCloser(&buf)

		start := time.Now()
		h.ServeHTTP(rw, r)
		duration := time.Since(start)

		l.Log.Infof("HTTP request method: %s, headers: %v, body: %s, url: %s, duration: %s, statusCode: %d, responseSize: %d",
			r.Method, r.Header, string(body), r.RequestURI, duration, rw.responseData.status, rw.responseData.size,
		)
	})
}

func (l *AppLogger) Interrupt() error {

	if err := l.Log.Sync(); err != nil {

		if runtime.GOOS == "darwin" {
			return nil
		} else {
			return fmt.Errorf("cannot flush buffered log entries err: %w", err)
		}

	}

	return nil

}

func (l *AppLogger) Info(template string, args ...interface{}) {
	l.Log.Infof(template, args)
}

func (l *AppLogger) Error(template string, args ...interface{}) {
	l.Log.Errorf(template, args)
}

func (l *AppLogger) Debug(template string, args ...interface{}) {
	l.Log.Debugf(template, args)
}

func (l *AppLogger) Warn(template string, args ...interface{}) {
	l.Log.Warnf(template, args)
}
