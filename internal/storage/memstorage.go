package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

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

func (ms *MemStorage) GetAllDataInt64(_ context.Context) (map[string]int64, error) {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataInt64, nil

}

func (ms *MemStorage) GetAllDataFloat64(_ context.Context) (map[string]float64, error) {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	return ms.dataFloat64, nil

}

func (ms *MemStorage) GetDataList(ctx context.Context) ([]string, error) {

	var list []string

	AllDataFloat64, err := ms.GetAllDataFloat64(ctx)
	if err != nil {
		return list, err
	}

	for k, v := range AllDataFloat64 {
		fv := strconv.FormatFloat(v, 'G', 12, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	AllDataInt64, err := ms.GetAllDataInt64(ctx)
	if err != nil {
		return list, err
	}

	for k, v := range AllDataInt64 {
		iv := strconv.FormatInt(v, 10)
		list = append(list, fmt.Sprintf("%s %s", k, iv))
	}

	return list, nil

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

	AllDataFloat64, err := ms.GetAllDataFloat64(ctx)
	if err != nil {
		return nil, err
	}

	float64map := make(map[string]string)
	for k, v := range AllDataFloat64 {
		fv := strconv.FormatFloat(v, 'G', 18, 64)
		float64map[k] = fv
	}

	AllDataInt64, err := ms.GetAllDataInt64(ctx)
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

	return json.Marshal(state)

}

func (ms *MemStorage) BatchSetFloat64Value(_ context.Context, gauges map[string]float64) (map[string]float64, []error, error) {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for key, value := range gauges {
		v, ok := ms.dataFloat64[key]
		if !ok {
			v = 0
		}
		newValue := v + value
		ms.dataFloat64[key] = newValue
	}
	var errs []error
	return gauges, errs, nil
}
func (ms *MemStorage) BatchAddInt64Value(_ context.Context, counters map[string]int64) (map[string]int64, []error, error) {

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
