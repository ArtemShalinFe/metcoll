package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type AppLogger struct {
	*zap.SugaredLogger
}

func NewLogger() (*AppLogger, error) {

	l, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("cannot init zap-logger err: %v ", err)
	}

	sl := l.Sugar()

	return &AppLogger{
		sl,
	}, nil

}

func (l *AppLogger) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := NewResponseLoggerWriter(w)

		start := time.Now()
		h.ServeHTTP(rw, r)
		duration := time.Since(start)

		l.Info("incomming HTTP request - ",
			"method:", r.Method,
			", path:", r.RequestURI,
			", duration:", duration,
			", statusCode:", rw.responseData.status,
			", responseSize:", rw.responseData.size,
		)
	})
}

func (l *AppLogger) LoggerInterrupt() error {

	if err := l.Sync(); err != nil {
		return fmt.Errorf("cannot flush buffered log entries err: %v", err)
	}

	return nil

}
