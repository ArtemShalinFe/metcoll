package logger

import (
	"go.uber.org/zap"
)

type RetryHTTPLogger struct {
	*zap.SugaredLogger
}

func NewRLLogger(l *zap.SugaredLogger) (*RetryHTTPLogger, error) {

	return &RetryHTTPLogger{
		l,
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
