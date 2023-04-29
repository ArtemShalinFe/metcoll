package storage

var Values *MemStorage

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: make(map[string]float64),
	}
}

type Storage interface {
	Get(Key string) interface{}
	Set(Key string, Value interface{})
}

type MemStorage struct {
	storage map[string]float64
}

func (ms *MemStorage) Get(Key string) float64 {
	return Values.storage[Key]
}

func (ms *MemStorage) Set(Key string, Value float64) {
	Values.storage[Key] = Value
}

func init() {
	Values = NewMemStorage()
}
