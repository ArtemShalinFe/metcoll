package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
)

type MemStorage struct {
	mutex       *sync.Mutex
	dataInt64   map[string]int64
	dataFloat64 map[string]float64
}

type Storage interface {
	GetInt64Value(ctx context.Context, key string) (int64, error)
	GetFloat64Value(ctx context.Context, key string) (float64, error)
	AddInt64Value(ctx context.Context, key string, value int64) (int64, error)
	SetFloat64Value(ctx context.Context, key string, value float64) (float64, error)
	GetDataList(ctx context.Context) ([]string, error)
	BatchSetFloat64Value(ctx context.Context, gauges map[string]float64) (map[string]float64, []error, error)
	BatchAddInt64Value(ctx context.Context, counters map[string]int64) (map[string]int64, []error, error)
	Interrupt() error
	Ping(ctx context.Context) error
}

var ErrNoRows = errors.New("no rows in result")

func InitStorage(ctx context.Context, cfg *configuration.Config, l *zap.SugaredLogger) (Storage, error) {

	pctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if strings.TrimSpace(cfg.Database) != "" {

		db, err := newSQLStorage(pctx, cfg.Database, l)
		if err != nil {
			return nil, fmt.Errorf("cannot init db storage err: %s", err)
		}

		return db, nil

	} else if strings.TrimSpace(cfg.FileStoragePath) != "" {

		fs, err := newFilestorage(newMemStorage(), l, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
		if err != nil {
			return nil, fmt.Errorf("cannot init filestorage err: %s", err)
		}

		return fs, nil

	} else {

		l.Info("saving the state to a filestorage has been disabled - empty filestorage path")
		return newMemStorage(), nil

	}

}
