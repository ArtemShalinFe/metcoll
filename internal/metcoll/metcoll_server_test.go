package metcoll

import (
	"context"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"go.uber.org/zap"
)

func TestInitServer(t *testing.T) {
	ctx := context.Background()
	logger := zap.S()

	cfgHTTP := &configuration.Config{}
	wantHTTPServer, err := NewHTTPServer(ctx, nil, cfgHTTP, logger)
	if err != nil {
		t.Errorf("init http server, err: %v", err)
	}

	cfgGRPC := &configuration.Config{}
	cfgGRPC.UseProtobuff = true

	wantGRPCServer, err := NewGRPCServer(nil, cfgHTTP, logger)
	if err != nil {
		t.Errorf("init grpc server, err: %v", err)
	}
	tests := []struct {
		name    string
		cfg     *configuration.Config
		want    MetricServer
		wantErr bool
	}{
		{
			name:    "http server",
			cfg:     cfgHTTP,
			want:    wantHTTPServer,
			wantErr: false,
		},
		{
			name:    "grpc server",
			cfg:     cfgGRPC,
			want:    wantGRPCServer,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitServer(ctx, nil, tt.cfg, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// tp := tt.want.(type)
			switch tp := got.(type) {
			case *HTTPServer:
				_, ok := tt.want.(*HTTPServer)
				if !ok {
					t.Errorf("wrong type, got %v, want HTTPServer type", tp)
				}
			case *GRPCServer:
				_, ok := tt.want.(*GRPCServer)
				if !ok {
					t.Errorf("wrong type, got %v, want GRPCServer type", tp)
				}
			default:
				t.Error("unknow type")
			}
		})
	}
}

// func TestInitClient(t *testing.T) {
// 	ctx := context.Background()
// 	logger := zap.S()

// 	httpCfg := &configuration.ConfigAgent{}
// 	wantHTTPClient, err := NewHTTPClient(httpCfg, logger)
// 	if err != nil {
// 		t.Errorf("init http client, err: %v", err)
// 	}
// 	tests := []struct {
// 		name    string
// 		cfg     *configuration.ConfigAgent
// 		want    MetricUpdater
// 		wantErr bool
// 	}{
// 		{
// 			name:    "http client",
// 			cfg:     httpCfg,
// 			want:    wantHTTPClient,
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := InitClient(ctx, tt.cfg, logger)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("InitClient() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			// tp := tt.want.(type)
// 			switch tp := got.(type) {
// 			case *Client:
// 				_, ok := tt.want.(*Client)
// 				if !ok {
// 					t.Errorf("wrong type, got %v, want HTTPClient type", tp)
// 				}
// 			default:
// 				t.Error("unknow type")
// 			}
// 		})
// 	}
// }
