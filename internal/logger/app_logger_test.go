package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestMiddlewareLogger_Interrupt(t *testing.T) {
	type fields struct {
		SugaredLogger *zap.SugaredLogger
	}
	tests := []struct {
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				SugaredLogger: zap.L().Sugar(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &MiddlewareLogger{
				SugaredLogger: tt.fields.SugaredLogger,
			}
			if err := l.Interrupt(); (err != nil) != tt.wantErr {
				t.Errorf("MiddlewareLogger.Interrupt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
