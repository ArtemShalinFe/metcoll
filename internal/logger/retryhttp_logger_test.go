package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewRLLogger(t *testing.T) {
	l := zap.L().Sugar()
	lg, err := NewRLLogger(l)
	if err != nil {
		t.Error(err)
	}

	const testMsg = "test"

	lg.Error(testMsg)
	lg.Debug(testMsg)
	lg.Info(testMsg)
	lg.Warn(testMsg)
}
