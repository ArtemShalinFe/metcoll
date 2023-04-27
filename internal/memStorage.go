package internal

var Values *MemStorage

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: make(map[string]uint64),
	}
}

type MemStorage struct {
	storage map[string]uint64
}

func (ms *MemStorage) Get(Key string) uint64 {
	return Values.storage[Key]
}

func (ms *MemStorage) Set(Key string, Value uint64) {
	Values.storage[Key] = Value
}

func init() {
	Values = NewMemStorage()
}
