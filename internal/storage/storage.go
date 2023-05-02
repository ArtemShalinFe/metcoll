package storage

import "errors"

var ErrWrongType = errors.New("wrong value type")

var Values *MemStorage

type MemStorage struct {
	Data map[string]interface{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Data: make(map[string]interface{}),
	}
}

func (ms *MemStorage) StorageHaveMetric(key string) bool {

	_, have := ms.Data[key]
	return have
}

func (ms *MemStorage) GetInt64Value(key string) (int64, error) {

	v := ms.Data[key]
	switch v.(type) {
	case int64:
		return ms.Data[key].(int64), nil
	case nil:
		return 0, nil
	default:
		var def int64
		return def, ErrWrongType
	}
}

func (ms *MemStorage) GetFloat64Value(key string) (float64, error) {

	switch ms.Data[key].(type) {
	case float64:
		return ms.Data[key].(float64), nil
	case nil:
		return 0, nil
	default:
		var def float64
		return def, ErrWrongType
	}
}

func (ms *MemStorage) SetInt64Value(key string, value int64) error {

	ms.Data[key] = value

	return nil
}

func (ms *MemStorage) SetFloat64Value(key string, value float64) error {

	ms.Data[key] = value

	return nil
}

func (ms *MemStorage) GetMetricList() map[string]interface{} {
	return Values.Data
}

func init() {
	Values = NewMemStorage()
}
