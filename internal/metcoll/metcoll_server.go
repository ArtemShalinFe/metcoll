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

type Storage interface {
	// GetInt64Value - returns the metric value or ErrNoRows if it does not exist.
	GetInt64Value(ctx context.Context, key string) (int64, error)

	// GetFloat64Value - returns the metric value or ErrNoRows if it does not exist.
	GetFloat64Value(ctx context.Context, key string) (float64, error)

	// AddInt64Value - Saves the metric value for the key and returns the new metric value.
	AddInt64Value(ctx context.Context, key string, value int64) (int64, error)

	// SetFloat64Value - Saves the metric value for the key and returns the new metric value.
	SetFloat64Value(ctx context.Context, key string, value float64) (float64, error)

	// GetDataList - Returns all saved metrics.
	// Metric output format: <MetricName> <Value>
	//
	// Example:
	//
	//	MetricOne 1
	//	MetricTwo 2
	//	...
	GetDataList(ctx context.Context) ([]string, error)

	// BatchSetFloat64Value - Batch saving of metric values.
	// Returns the set metric values and errors for those metrics whose values could not be set.
	BatchSetFloat64Value(ctx context.Context, gauges map[string]float64) (map[string]float64, []error, error)

	// BatchAddInt64Value - Batch saving of metric values.
	// Returns the set metric values and errors for those metrics whose values could not be set.
	BatchAddInt64Value(ctx context.Context, counters map[string]int64) (map[string]int64, []error, error)

	Ping(ctx context.Context) error
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
