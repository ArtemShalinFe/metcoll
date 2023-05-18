package storage

import (
	"fmt"
	"strconv"
	"sync"
)

type MemStorage struct {
	mutex       *sync.Mutex
	dataInt64   map[string]int64
	dataFloat64 map[string]float64
}

func NewMemStorage() *MemStorage {

	return &MemStorage{
		mutex:       &sync.Mutex{},
		dataInt64:   make(map[string]int64),
		dataFloat64: make(map[string]float64),
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

func (ms *MemStorage) GetCounterList() []string {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	var list []string

	for k, v := range ms.dataInt64 {
		iv := strconv.FormatInt(v, 10)
		list = append(list, fmt.Sprintf("%s %s", k, iv))
	}

	return list

}

func (ms *MemStorage) GetGaugeList() []string {

	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	var list []string

	for k, v := range ms.dataFloat64 {
		fv := strconv.FormatFloat(v, 'G', 10, 64)
		list = append(list, fmt.Sprintf("%s %s", k, fv))
	}

	return list

}

func (ms *MemStorage) GetDataList() []string {

	var list []string

	list = append(list, ms.GetGaugeList()...)
	list = append(list, ms.GetCounterList()...)

	return list

}
