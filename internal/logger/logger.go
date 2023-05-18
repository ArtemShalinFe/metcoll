package logger

import (
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
		return nil, err
	}
	defer l.Sync()

	sl := l.Sugar()

	return &AppLogger{
		sl,
	}, nil

}

func (l *AppLogger) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		rw := NewResponseLoggerWriter(w)

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
