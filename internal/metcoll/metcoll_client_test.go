package metcoll

import (
	"context"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"go.uber.org/zap"
)

func TestInitClient(t *testing.T) {
	ctx := context.Background()
	logger := zap.S()

	cfgHTTP := &configuration.ConfigAgent{}
	wantHTTPClient, err := NewHTTPClient(cfgHTTP, logger)
	if err != nil {
		t.Errorf("init http client, err: %v", err)
	}
	tests := []struct {
		want    MetricUpdater
		cfg     *configuration.ConfigAgent
		name    string
		wantErr bool
	}{
		{
			name:    "http client",
			cfg:     cfgHTTP,
			want:    wantHTTPClient,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitClient(ctx, tt.cfg, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// tp := tt.want.(type)
			switch tp := got.(type) {
			case *Client:
				_, ok := tt.want.(*Client)
				if !ok {
					t.Errorf("wrong type, got %v, want HTTPClient type", tp)
				}
			default:
				t.Error("unknow type")
			}
		})
	}
}
