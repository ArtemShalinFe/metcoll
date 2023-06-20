package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type RetryHTTPLogger struct {
	*zap.SugaredLogger
}

func NewRLLogger() (*RetryHTTPLogger, error) {

	l, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("cannot init zap-logger err: %w ", err)
	}

	sl := l.Sugar()

	return &RetryHTTPLogger{
		sl,
	}, nil

}
func (rl *RetryHTTPLogger) Error(msg string, keysAndValues ...interface{}) {
	rl.Errorf(msg, keysAndValues...)
}

func (rl *RetryHTTPLogger) Info(msg string, keysAndValues ...interface{}) {
	rl.Infof(msg, keysAndValues...)
}
func (rl *RetryHTTPLogger) Debug(msg string, keysAndValues ...interface{}) {
	rl.Debugf(msg, keysAndValues...)
}
func (rl *RetryHTTPLogger) Warn(msg string, keysAndValues ...interface{}) {
	rl.Warnf(msg, keysAndValues...)
}
