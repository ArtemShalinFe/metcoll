package storage

import (
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
	GetInt64Value(key string) (int64, bool)
	GetFloat64Value(key string) (float64, bool)
	AddInt64Value(key string, value int64) int64
	SetFloat64Value(key string, value float64) float64
	GetDataList() []string
	Interrupt() error
}

func NewMemStorage() *MemStorage {

	ms := &MemStorage{
		mutex:       &sync.Mutex{},
		dataInt64:   make(map[string]int64),
		dataFloat64: make(map[string]float64),
	}

	return ms

}

func (ms *MemStorage) GetInt64Value(key string) (int64, bool) {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataInt64[key]
	return v, ok

}

func (ms *MemStorage) GetFloat64Value(key string) (float64, bool) {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataFloat64[key]
	return v, ok

}

func (ms *MemStorage) AddInt64Value(key string, value int64) int64 {

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

func (ms *MemStorage) SetFloat64Value(key string, value float64) float64 {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.dataFloat64[key] = value

	return value

}

func (ms *MemStorage) GetAllDataInt64() map[string]int64 {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataInt64

}

func (ms *MemStorage) GetAllDataFloat64() map[string]float64 {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataFloat64

}

func (ms *MemStorage) GetDataList() []string {

	var list []string

	for k, v := range ms.GetAllDataFloat64() {
		fv := strconv.FormatFloat(v, 'G', 12, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	for k, v := range ms.GetAllDataInt64() {
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

	stateFloat64 := state["float64"]
	for k, v := range stateFloat64 {
		pv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		ms.SetFloat64Value(k, pv)
	}

	stateInt64 := state["int64"]
	for k, v := range stateInt64 {
		pv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		ms.AddInt64Value(k, pv)
	}

	return nil

}

func (ms *MemStorage) MarshalJSON() ([]byte, error) {

	float64map := make(map[string]string)
	for k, v := range ms.GetAllDataFloat64() {
		fv := strconv.FormatFloat(v, 'G', 18, 64)
		float64map[k] = fv
	}

	int64map := make(map[string]string)
	for k, v := range ms.GetAllDataInt64() {
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

func InitStorage(cfg *configuration.Config, s *MemStorage, l Logger) (Storage, error) {

	if strings.TrimSpace(cfg.FileStoragePath) != "" {

		fs, err := newFilestorage(s, l, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
		if err != nil {
			l.Error("cannot init filestorage err: ", err)
			return nil, err
		}

		return fs, nil

	}

	l.Info("saving the state to a filestorage has been disabled - empty filestorage path")
	return s, nil

}
