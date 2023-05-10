package storage

import (
	"sync"
)

type MemStorage struct {
	sync.Mutex
	data map[string]interface{}
}

func NewMemStorage() *MemStorage {

	return &MemStorage{
		data: make(map[string]interface{}),
	}

}

func (ms *MemStorage) GetInt64Value(key string) (int64, bool) {

	v, ok := ms.data[key].(int64)
	return v, ok

}

func (ms *MemStorage) GetFloat64Value(key string) (float64, bool) {

	v, ok := ms.data[key].(float64)
	return v, ok

}

func (ms *MemStorage) AddInt64Value(key string, value int64) int64 {

	ms.Lock()

	v, ok := ms.data[key].(int64)
	if !ok {
		v = 0
	}
	newValue := v + value
	ms.data[key] = newValue

	ms.Unlock()

	return newValue

}

func (ms *MemStorage) SetFloat64Value(key string, value float64) float64 {

	ms.Lock()
	ms.data[key] = value
	ms.Unlock()

	return value

}

func (ms *MemStorage) GetDataList() map[string]interface{} {

	return ms.data

}
