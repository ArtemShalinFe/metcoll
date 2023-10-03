package metcoll

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type MetricUpdater interface {
	BatchUpdateMetric(ctx context.Context, mcs <-chan []*metrics.Metrics, result chan<- error)
}

func InitClient(ctx context.Context, cfg *configuration.ConfigAgent, sl *zap.SugaredLogger) (MetricUpdater, error) {
	if cfg.UseProtobuff {
		grpcClient, err := NewGRPCClient(ctx, cfg, sl)
		if err != nil {
			return nil, fmt.Errorf("an occured error when init grpc server, err: %w", err)
		}
		return grpcClient, nil
	} else {
		httpClient, err := NewHTTPClient(cfg, sl)
		if err != nil {
			return nil, fmt.Errorf("an occured error when init http server, err: %w", err)
		}
		return httpClient, nil
	}
}
