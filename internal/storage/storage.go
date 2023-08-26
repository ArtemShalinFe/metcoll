// The package describes the interaction of the server with various sources of metrics storage.
// Metrics can be stored in memory, in a file on disk, in a database.
package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
)

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

	// Interrupt - function for gracefull shutdown.
	Interrupt() error

	Ping(ctx context.Context) error
}

// ErrNoRows - if the metric is not found in the repository, the service will show this error.
var ErrNoRows = errors.New("no rows in result")

func InitStorage(ctx context.Context, cfg *configuration.Config, l *zap.SugaredLogger) (Storage, error) {
	pctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if strings.TrimSpace(cfg.Database) != "" {
		db, err := newSQLStorage(pctx, cfg.Database, l)
		if err != nil {
			return nil, fmt.Errorf("cannot init db storage err: %w", err)
		}

		return db, nil
	} else if strings.TrimSpace(cfg.FileStoragePath) != "" {
		fs, err := newFilestorage(newMemStorage(), l, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
		if err != nil {
			return nil, fmt.Errorf("cannot init filestorage err: %w", err)
		}

		return fs, nil
	} else {
		l.Info("saving the state to a filestorage has been disabled - empty filestorage path")
		return newMemStorage(), nil
	}
}
