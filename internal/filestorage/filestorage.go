package filestorage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type Filestorage struct {
	*storage.MemStorage
	path          string
	storeInterval int
	logger        Logger
}

func NewFilestorage(stg *storage.MemStorage, l Logger, path string, storeInterval int, restore bool) (*Filestorage, error) {

	fs := &Filestorage{
		MemStorage:    stg,
		logger:        l,
		path:          path,
		storeInterval: storeInterval,
	}

	if restore {
		if err := fs.Load(fs.MemStorage); err != nil {
			fs.logger.Info("cannot restore state storage err: ", err)
			return fs, nil
		}
	}

	if storeInterval > 0 {
		go fs.runIntervalStateSaving()
	}

	return fs, nil

}

func (fs *Filestorage) AddInt64Value(key string, value int64) int64 {

	newValue := fs.MemStorage.AddInt64Value(key, value)
	if fs.storeInterval == 0 {
		fs.Save(fs.MemStorage)
	}
	return newValue

}

func (fs *Filestorage) SetFloat64Value(key string, value float64) float64 {

	newValue := fs.MemStorage.SetFloat64Value(key, value)
	if fs.storeInterval == 0 {
		fs.Save(fs.MemStorage)
	}
	return newValue

}

func (fs *Filestorage) Save(storage *storage.MemStorage) error {

	file, err := os.OpenFile(fs.path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fs.logger.Error("cannot open or creating file for state saving err: ", err)
		return err
	}
	defer file.Close()

	fs.logger.Info("try save state")

	data, err := storage.GetState()
	if err != nil {
		fs.logger.Error("cannot get storage state err: ", err)
		return err
	}

	data = append(data, '\n')

	if _, err = file.Write(data); err != nil {
		return nil
	}

	fs.logger.Info("state was saved")

	return nil

}

func (fs *Filestorage) Load(storage *storage.MemStorage) error {

	file, err := os.OpenFile(fs.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		fs.logger.Error("cannot open or creating file for state loading err: ", err)
		return err
	}

	fs.logger.Info("try restoring state")

	defer file.Close()

	r := bufio.NewReader(file)
	b, err := r.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			fs.logger.Error("cannot reading file for state loading err: ", err)
			return err
		}
	}

	if len(b) == 0 {
		fs.logger.Info("state cannot be restored - file is empty")
		return nil
	}

	if err = storage.SetState(b); err != nil {
		fs.logger.Error("cannot set state err: ", err)
		return err
	}

	fs.logger.Info("storage state was restored")

	return nil

}

func (fs *Filestorage) runIntervalStateSaving() {

	sleepDuration := time.Duration(fs.storeInterval) * time.Second

	for {
		time.Sleep(sleepDuration)
		if err := fs.Save(fs.MemStorage); err != nil {
			fs.logger.Error("cannot save state err: ", err)
		}
	}

}

func (fs *Filestorage) FilestorageInterrupt() error {

	if err := fs.Save(fs.MemStorage); err != nil {
		return fmt.Errorf("cannot save state err: %v", err)
	}

	return nil
}
