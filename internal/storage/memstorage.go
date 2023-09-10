package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

// MemStorage - implementation of a in-memory database for storing metrics.
type MemStorage struct {
	mutex       *sync.Mutex
	dataInt64   map[string]int64
	dataFloat64 map[string]float64
}

func newMemStorage() *MemStorage {
	ms := &MemStorage{
		mutex:       &sync.Mutex{},
		dataInt64:   make(map[string]int64),
		dataFloat64: make(map[string]float64),
	}

	return ms
}

func (ms *MemStorage) GetInt64Value(_ context.Context, key string) (int64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataInt64[key]
	if !ok {
		return 0, ErrNoRows
	}

	return v, nil
}

func (ms *MemStorage) GetFloat64Value(_ context.Context, key string) (float64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataFloat64[key]
	if !ok {
		return 0, ErrNoRows
	}
	return v, nil
}

func (ms *MemStorage) AddInt64Value(_ context.Context, key string, value int64) (int64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	v, ok := ms.dataInt64[key]
	if !ok {
		v = 0
	}
	newValue := v + value
	ms.dataInt64[key] = newValue

	return newValue, nil
}

func (ms *MemStorage) SetFloat64Value(_ context.Context, key string, value float64) (float64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.dataFloat64[key] = value

	return value, nil
}

func (ms *MemStorage) getAllDataInt64(_ context.Context) (map[string]int64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataInt64, nil
}

func (ms *MemStorage) getAllDataFloat64(_ context.Context) (map[string]float64, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataFloat64, nil
}

func (ms *MemStorage) GetDataList(ctx context.Context) ([]string, error) {
	AllDataFloat64, err := ms.getAllDataFloat64(ctx)
	if err != nil {
		return nil, err
	}
	AllDataInt64, err := ms.getAllDataInt64(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0, len(AllDataFloat64)+len(AllDataInt64))

	for k, v := range AllDataFloat64 {
		fv := strconv.FormatFloat(v, 'G', 12, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	for k, v := range AllDataInt64 {
		iv := strconv.FormatInt(v, 10)
		list = append(list, fmt.Sprintf("%s %s", k, iv))
	}

	return list, nil
}

func (ms *MemStorage) GetState() ([]byte, error) {
	b, err := json.Marshal(&ms)
	if err != nil {
		return nil, fmt.Errorf("memory storage get state err: %w", err)
	}
	return b, nil
}

func (ms *MemStorage) SetState(data []byte) error {
	if err := json.Unmarshal(data, &ms); err != nil {
		return fmt.Errorf("memory storage set state err: %w", err)
	}
	return nil
}

func (ms *MemStorage) UnmarshalJSON(b []byte) error {
	state := make(map[string]map[string]string)

	if err := json.Unmarshal(b, &state); err != nil {
		return fmt.Errorf("memory storage unmarshal state err: %w", err)
	}

	ctx := context.Background()

	stateFloat64 := state["float64"]
	for k, v := range stateFloat64 {
		pv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("parse float memory storage err: %w", err)
		}
		if _, err := ms.SetFloat64Value(ctx, k, pv); err != nil {
			return fmt.Errorf("memory storage set float err: %w", err)
		}
	}

	stateInt64 := state["int64"]
	for k, v := range stateInt64 {
		pv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int memory storage err: %w", err)
		}

		if _, err := ms.AddInt64Value(ctx, k, pv); err != nil {
			return fmt.Errorf("memory storage set int err: %w", err)
		}
	}

	return nil
}

func (ms *MemStorage) MarshalJSON() ([]byte, error) {
	ctx := context.Background()

	AllDataFloat64, err := ms.getAllDataFloat64(ctx)
	if err != nil {
		return nil, err
	}

	float64map := make(map[string]string)
	for k, v := range AllDataFloat64 {
		fv := strconv.FormatFloat(v, 'G', 18, 64)
		float64map[k] = fv
	}

	AllDataInt64, err := ms.getAllDataInt64(ctx)
	if err != nil {
		return nil, err
	}
	int64map := make(map[string]string)
	for k, v := range AllDataInt64 {
		iv := strconv.FormatInt(v, 10)
		int64map[k] = iv
	}

	state := make(map[string]map[string]string)
	state["float64"] = float64map
	state["int64"] = int64map

	b, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("memory storage marshal err: %w", err)
	}

	return b, nil
}

func (ms *MemStorage) BatchSetFloat64Value(_ context.Context,
	gauges map[string]float64) (map[string]float64, []error, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for key, value := range gauges {
		ms.dataFloat64[key] = value
	}
	var errs []error
	return gauges, errs, nil
}

func (ms *MemStorage) BatchAddInt64Value(_ context.Context,
	counters map[string]int64) (map[string]int64, []error, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for key, value := range counters {
		v, ok := ms.dataInt64[key]
		if !ok {
			v = 0
		}
		newValue := v + value
		ms.dataInt64[key] = newValue
	}

	var errs []error
	return counters, errs, nil
}

func (ms *MemStorage) Interrupt() error {
	return nil
}

func (ms *MemStorage) Ping(ctx context.Context) error {
	return nil
}
