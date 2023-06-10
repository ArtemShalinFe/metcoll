package storage

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/sleepstepper"
)

type Logger interface {
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

type Filestorage struct {
	*MemStorage
	path          string
	storeInterval int
	logger        Logger
}

func newFilestorage(stg *MemStorage, l Logger, path string, storeInterval int, restore bool) (*Filestorage, error) {

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

func (fs *Filestorage) AddInt64Value(ctx context.Context, key string, value int64) int64 {

	newValue := fs.MemStorage.AddInt64Value(ctx, key, value)
	if fs.storeInterval == 0 {
		if err := fs.Save(fs.MemStorage); err != nil {
			fs.logger.Errorf("synchronous saving to file storage cannot be performed err: %w", err)
		}
	}
	return newValue

}

func (fs *Filestorage) SetFloat64Value(ctx context.Context, key string, value float64) float64 {

	newValue := fs.MemStorage.SetFloat64Value(ctx, key, value)
	if fs.storeInterval == 0 {
		if err := fs.Save(fs.MemStorage); err != nil {
			fs.logger.Errorf("synchronous saving to file storage cannot be performed err: %w", err)
		}
	}
	return newValue

}

func (fs *Filestorage) Save(storage *MemStorage) error {

	ss := sleepstepper.NewSleepStepper(1, 2, 5)
	file, err := retryOpenFile(os.OpenFile, ss, fs.path, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return fmt.Errorf("cannot open or creating file for state saving err: %w", err)
	}
	defer file.Close()

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

	ss := sleepstepper.NewSleepStepper(1, 2, 5)
	file, err := retryOpenFile(os.OpenFile, ss, fs.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("cannot open or creating file for state loading err: %w", err)
	}

	fs.logger.Info("try restoring state")

	defer file.Close()

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
			fs.logger.Errorf("cannot save state err: %w", err)
		}
	}

}

func (fs *Filestorage) Interrupt() error {

	if err := fs.Save(fs.MemStorage); err != nil {
		return fmt.Errorf("cannot save state err: %w", err)
	}

	return nil
}

func (fs *Filestorage) Ping(ctx context.Context) error {
	return nil
}

type OpenFileFunc func(name string, flag int, perm fs.FileMode) (*os.File, error)

func retryOpenFile(f OpenFileFunc, ss Sleeper, name string, flag int, perm fs.FileMode) (*os.File, error) {

	row, err := f(name, flag, perm)
	if err != nil {

		if !errors.Is(err, syscall.EACCES) {
			return nil, err
		}

		if !ss.Sleep() {
			return nil, err
		}

		return retryOpenFile(f, ss, name, flag, perm)

	}

	return row, nil

}
