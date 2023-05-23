package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"
)

type stateSaver interface {
	Save(data []byte) error
	Load() ([]byte, error)
}

type MemStorage struct {
	mutex       *sync.Mutex
	dataInt64   map[string]int64
	dataFloat64 map[string]float64
	stateSaver  stateSaver
	syncSaving  bool
}

func NewMemStorage(restore bool, storeInterval int, stateSaver stateSaver) (*MemStorage, error) {

	ms := &MemStorage{
		mutex:       &sync.Mutex{},
		dataInt64:   make(map[string]int64),
		dataFloat64: make(map[string]float64),
		stateSaver:  stateSaver,
	}

	if restore {
		if err := ms.LoadState(); err != nil {
			return nil, err
		}
	}

	setupStateSaving(ms, storeInterval)

	return ms, nil

}

func setupStateSaving(ms *MemStorage, storeInterval int) {

	if storeInterval == 0 {
		ms.syncSaving = true
	} else {
		ms.syncSaving = false
		go ms.RunIntervalStateSaving(storeInterval)
	}

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

func (ms *MemStorage) AddInt64Value(key string, value int64) (int64, error) {

	newValue := ms.updateInt64Value(key, value)
	if ms.syncSaving {
		return newValue, ms.SaveState()
	} else {
		return newValue, nil
	}

}

func (ms *MemStorage) updateInt64Value(key string, value int64) int64 {

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

func (ms *MemStorage) SetFloat64Value(key string, value float64) (float64, error) {

	newValue := ms.updateFloat64Value(key, value)

	if ms.syncSaving {
		return newValue, ms.SaveState()
	} else {
		return newValue, nil
	}

}

func (ms *MemStorage) updateFloat64Value(key string, value float64) float64 {

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

func (ms *MemStorage) SaveState() error {

	storageState, err := json.Marshal(&ms)
	if err != nil {
		return err
	}

	return ms.stateSaver.Save(storageState)

}

func (ms *MemStorage) LoadState() error {

	data, err := ms.stateSaver.Load()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		} else {
			return err
		}
	}

	err = json.Unmarshal(data, &ms)
	if err != nil {
		return err
	}

	return nil

}

func (ms *MemStorage) RunIntervalStateSaving(storeInterval int) {

	sleepDuration := time.Duration(storeInterval) * time.Second
	for {
		if err := ms.SaveState(); err != nil {
			log.Printf("cannot save state err: %v\n", err)
		}
		time.Sleep(sleepDuration)
	}

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
