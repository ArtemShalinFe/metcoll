package storage

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
)

// Filestorage - implementation of a filestorage for interval saving the metrics.
type Filestorage struct {
	*MemStorage
	logger        *zap.SugaredLogger
	path          string
	storeInterval int
}

func newFilestorage(stg *MemStorage,
	l *zap.SugaredLogger,
	path string,
	storeInterval int,
	restore bool) (*Filestorage, error) {
	fs := &Filestorage{
		MemStorage:    stg,
		logger:        l,
		path:          path,
		storeInterval: storeInterval,
	}

	if restore {
		if err := fs.Load(fs.MemStorage); err != nil {
			fs.logger.Infof("cannot restore state storage err: %w", err)
			return fs, nil
		}
	}

	if storeInterval > 0 {
		go fs.runIntervalStateSaving()
	}

	return fs, nil
}

func (fs *Filestorage) BatchSetFloat64Value(ctx context.Context,
	gauges map[string]float64) (map[string]float64, []error, error) {
	gauges, errs, err := fs.MemStorage.BatchSetFloat64Value(ctx, gauges)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot add batch int64 value in filestorage err: %w", err)
	}

	if fs.storeInterval == 0 {
		if err := fs.Save(fs.MemStorage); err != nil {
			return nil, nil, fmt.Errorf("sync saving batch float64 value to file cannot be performed err: %w", err)
		}
	}

	return gauges, errs, nil
}

func (fs *Filestorage) BatchAddInt64Value(ctx context.Context,
	counters map[string]int64) (map[string]int64, []error, error) {
	counters, errs, err := fs.MemStorage.BatchAddInt64Value(ctx, counters)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot add batch int64 value in filestorage err: %w", err)
	}

	if fs.storeInterval == 0 {
		if err := fs.Save(fs.MemStorage); err != nil {
			return nil, nil, fmt.Errorf("sync saving batch int64 value to file cannot be performed err: %w", err)
		}
	}

	return counters, errs, nil
}

func (fs *Filestorage) AddInt64Value(ctx context.Context, key string, value int64) (int64, error) {
	newValue, err := fs.MemStorage.AddInt64Value(ctx, key, value)
	if err != nil {
		return 0, fmt.Errorf("cannot add int64 value in filestorage err: %w", err)
	}

	if fs.storeInterval == 0 {
		if err := fs.Save(fs.MemStorage); err != nil {
			return 0, fmt.Errorf("synchronous saving int64 to file storage cannot be performed err: %w", err)
		}
	}
	return newValue, nil
}

func (fs *Filestorage) SetFloat64Value(ctx context.Context, key string, value float64) (float64, error) {
	newValue, err := fs.MemStorage.SetFloat64Value(ctx, key, value)
	if err != nil {
		return 0, fmt.Errorf("cannot set float64 value in filestorage err: %w", err)
	}

	if fs.storeInterval == 0 {
		if err := fs.Save(fs.MemStorage); err != nil {
			return 0, fmt.Errorf("synchronous saving float64 to file storage cannot be performed err: %w", err)
		}
	}
	return newValue, nil
}

func (fs *Filestorage) Save(storage *MemStorage) error {
	file, err := os.OpenFile(fs.path, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return fmt.Errorf("cannot open or creating file for state saving err: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fs.logger.Errorf("closing file for save filestorage was failed err: %w", err)
		}
	}()

	fs.logger.Info("try save state")

	data, err := storage.GetState()
	if err != nil {
		return fmt.Errorf("cannot get storage state err: %w", err)
	}

	data = append(data, '\n')

	if _, err = file.Write(data); err != nil {
		return fmt.Errorf("cannot save state in filestorage err: %w", err)
	}

	fs.logger.Info("state was saved")

	return nil
}

func (fs *Filestorage) Load(storage *MemStorage) error {
	file, err := os.OpenFile(fs.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("cannot open or creating file for state loading err: %w", err)
	}

	fs.logger.Info("try restoring state")

	defer func() {
		if err := file.Close(); err != nil {
			fs.logger.Errorf("closing file for load filestorage was failed err: %w", err)
		}
	}()

	r := bufio.NewReader(file)
	b, err := r.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("cannot reading file for state loading err: %w", err)
		}
	}

	if len(b) == 0 {
		fs.logger.Info("state cannot be restored - file is empty")
		return nil
	}

	if err = storage.SetState(b); err != nil {
		return fmt.Errorf("cannot set state err: %w", err)
	}

	fs.logger.Info("storage state was restored")

	return nil
}

func (fs *Filestorage) runIntervalStateSaving() {
	sleepDuration := time.Duration(fs.storeInterval) * time.Second

	for {
		time.Sleep(sleepDuration)
		if err := fs.Save(fs.MemStorage); err != nil {
			fs.logger.Errorf("interval state saving cannot save state err: %w", err)
		}
	}
}

func (fs *Filestorage) Interrupt() error {
	if err := fs.Save(fs.MemStorage); err != nil {
		return fmt.Errorf("interrupt cannot save state err: %w", err)
	}

	return nil
}

func (fs *Filestorage) Ping(ctx context.Context) error {
	return nil
}
