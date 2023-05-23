package storage

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStateSaver struct{}

func (tss *testStateSaver) Save(data []byte) error {
	return nil
}

func (tss *testStateSaver) Load() ([]byte, error) {
	return nil, nil
}

func TestNewMemStorage(t *testing.T) {
	want := &MemStorage{
		mutex:       &sync.Mutex{},
		dataInt64:   make(map[string]int64),
		dataFloat64: make(map[string]float64),
		stateSaver:  &testStateSaver{},
		syncSaving:  true,
	}

	t.Run("Test mem storage constructor", func(t *testing.T) {
		if got, _ := NewMemStorage(false, 0, &testStateSaver{}); !reflect.DeepEqual(got, want) {
			t.Errorf("NewMemStorage() = %v, want %v", got, want)
		}
	})
}

func TestMemStorage_GetFloat64Value(t *testing.T) {

	ts, err := NewMemStorage(false, 0, &testStateSaver{})
	if err != nil {
		t.Error(err)
		return
	}

	ts.SetFloat64Value("test1", 1.0)
	ts.SetFloat64Value("test2", 2.0)
	ts.SetFloat64Value("test3", 4.0)

	type args struct {
		Key string
	}
	tests := []struct {
		name   string
		args   args
		want   float64
		wantOk bool
	}{
		{name: "test get storage value 1 - positive case",
			args:   args{Key: "test1"},
			want:   1.0,
			wantOk: true,
		},
		{name: "test get storage value 2 - positive case",
			args:   args{Key: "test2"},
			want:   2.0,
			wantOk: true,
		},
		{name: "test get storage value 4 - positive case",
			args:   args{Key: "test3"},
			want:   4.0,
			wantOk: true,
		},
		{name: "test get storage value none - negative case",
			args:   args{Key: "test4"},
			want:   0.0,
			wantOk: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			value, ok := ts.GetFloat64Value(tt.args.Key)
			if value != tt.want {
				t.Errorf("MemStorage.GetFloat64Value() = %v, want %v", value, tt.want)
			}
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestMemStorage_GetInt64Value(t *testing.T) {

	ts, err := NewMemStorage(false, 0, &testStateSaver{})
	if err != nil {
		t.Error(err)
		return
	}

	ts.AddInt64Value("test1", 1)
	ts.AddInt64Value("test2", 2)
	ts.AddInt64Value("test3", 3)
	ts.AddInt64Value("test3", 3)

	type args struct {
		Key string
	}
	tests := []struct {
		name   string
		args   args
		want   int64
		wantOk bool
	}{
		{name: "test get storage value 1 - positive case",
			args:   args{Key: "test1"},
			want:   1,
			wantOk: true,
		},
		{name: "test get storage value 2 - positive case",
			args:   args{Key: "test2"},
			want:   2,
			wantOk: true,
		},
		{name: "test get storage value 4 - negative case",
			args:   args{Key: "test3"},
			want:   6,
			wantOk: true,
		},
		{name: "test get storage value 0 - negative case",
			args:   args{Key: "test9"},
			want:   0,
			wantOk: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			value, ok := ts.GetInt64Value(tt.args.Key)
			if value != tt.want {
				t.Errorf("MemStorage.GetInt64Value() = %v, want %v", value, tt.want)
			}
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestMemStorage_GetDataList(t *testing.T) {

	ts, err := NewMemStorage(false, 0, &testStateSaver{})
	if err != nil {
		t.Error(err)
		return
	}
	ts.SetFloat64Value("test1", 1.2)
	ts.AddInt64Value("test4", 5)

	var want []string
	want = append(want, "test1 1.2")
	want = append(want, "test4 5")

	tests := []struct {
		name string
		ms   *MemStorage
		want []string
	}{
		{
			name: "get data list",
			ms:   ts,
			want: want,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.ms.GetDataList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.GetDataList() = %v, want %v", got, tt.want)
			}
		})
	}
}
