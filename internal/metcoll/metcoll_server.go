package metcoll

import (
	"context"
	"fmt"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"go.uber.org/zap"
)

type MetcollServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

func InitServer(ctx context.Context, stg Storage, cfg *configuration.Config, sl *zap.SugaredLogger) (MetcollServer, error) {
	if cfg.UseProtobuff {
		grpcServer, err := NewGRPCServer(stg, cfg, sl)
		if err != nil {
			return nil, fmt.Errorf("an occured error when init grpc server, err: %w", err)
		}
		return grpcServer, nil
	} else {
		httpServer, err := NewHTTPServer(ctx, stg, cfg, sl)
		if err != nil {
			return nil, fmt.Errorf("an occured error when init http server, err: %w", err)
		}
		return httpServer, nil
	}
}
