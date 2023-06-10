package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
)

type MemStorage struct {
	mutex       *sync.Mutex
	dataInt64   map[string]int64
	dataFloat64 map[string]float64
}

type Storage interface {
	GetInt64Value(ctx context.Context, key string) (int64, bool)
	GetFloat64Value(ctx context.Context, key string) (float64, bool)
	AddInt64Value(ctx context.Context, key string, value int64) int64
	SetFloat64Value(ctx context.Context, key string, value float64) float64
	GetDataList(ctx context.Context) []string
	Interrupt() error
	Ping(ctx context.Context) error
}

func newMemStorage() *MemStorage {

	ms := &MemStorage{
		mutex:       &sync.Mutex{},
		dataInt64:   make(map[string]int64),
		dataFloat64: make(map[string]float64),
	}

	return ms

}

func (ms *MemStorage) GetInt64Value(ctx context.Context, key string) (int64, bool) {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataInt64[key]
	return v, ok

}

func (ms *MemStorage) GetFloat64Value(ctx context.Context, key string) (float64, bool) {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataFloat64[key]
	return v, ok

}

func (ms *MemStorage) AddInt64Value(ctx context.Context, key string, value int64) int64 {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataInt64[key]
	if !ok {
		v = 0
	}
	newValue := v + value
	ms.dataInt64[key] = newValue
	return newValue

}

func (ms *MemStorage) SetFloat64Value(ctx context.Context, key string, value float64) float64 {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.dataFloat64[key] = value

	return value

}

func (ms *MemStorage) GetAllDataInt64(ctx context.Context) map[string]int64 {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataInt64

}

func (ms *MemStorage) GetAllDataFloat64(ctx context.Context) map[string]float64 {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataFloat64

}

func (ms *MemStorage) GetDataList(ctx context.Context) []string {

	var list []string

	for k, v := range ms.GetAllDataFloat64(ctx) {
		fv := strconv.FormatFloat(v, 'G', 12, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	for k, v := range ms.GetAllDataInt64(ctx) {
		iv := strconv.FormatInt(v, 10)
		list = append(list, fmt.Sprintf("%s %s", k, iv))
	}

	return list

}

func (ms *MemStorage) GetState() ([]byte, error) {

	return json.Marshal(&ms)

}

func (ms *MemStorage) SetState(data []byte) error {

	return json.Unmarshal(data, &ms)

}

func (ms *MemStorage) UnmarshalJSON(b []byte) error {

	state := make(map[string]map[string]string)

	if err := json.Unmarshal(b, &state); err != nil {
		return err
	}

	ctx := context.Background()

	stateFloat64 := state["float64"]
	for k, v := range stateFloat64 {
		pv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		ms.SetFloat64Value(ctx, k, pv)
	}

	stateInt64 := state["int64"]
	for k, v := range stateInt64 {
		pv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		ms.AddInt64Value(ctx, k, pv)
	}

	return nil

}

func (ms *MemStorage) MarshalJSON() ([]byte, error) {

	ctx := context.Background()

	float64map := make(map[string]string)
	for k, v := range ms.GetAllDataFloat64(ctx) {
		fv := strconv.FormatFloat(v, 'G', 18, 64)
		float64map[k] = fv
	}

	int64map := make(map[string]string)
	for k, v := range ms.GetAllDataInt64(ctx) {
		iv := strconv.FormatInt(v, 10)
		int64map[k] = iv
	}

	state := make(map[string]map[string]string)
	state["float64"] = float64map
	state["int64"] = int64map

	return json.Marshal(state)

}

func (ms *MemStorage) Interrupt() error {
	return nil
}

func (ms *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func InitStorage(ctx context.Context, cfg *configuration.Config, l Logger) (Storage, error) {

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
