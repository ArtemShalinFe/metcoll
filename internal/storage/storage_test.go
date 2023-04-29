package storage

import (
	"reflect"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	want := &MemStorage{
		storage: make(map[string]float64),
	}

	t.Run("Test mem storage constructor", func(t *testing.T) {
		if got := NewMemStorage(); !reflect.DeepEqual(got, want) {
			t.Errorf("NewMemStorage() = %v, want %v", got, want)
		}
	})
}

func TestMemStorage_Get(t *testing.T) {
	ts := NewMemStorage()
	ts.Set("test1", 1.0)
	ts.Set("test2", 2.0)
	ts.Set("test3", 4.0)

	type args struct {
		Key string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "test get storage value 1 - positive case",
			args: args{Key: "test1"},
			want: 1.0,
		},
		{name: "test get storage value 2 - positive case",
			args: args{Key: "test2"},
			want: 2.0,
		},
		{name: "test get storage value 4 - positive case",
			args: args{Key: "test3"},
			want: 4.0,
		},
		{name: "test get storage value 0 - negative case",
			args: args{Key: "test8"},
			want: 0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ts.Get(tt.args.Key); got != tt.want {
				t.Errorf("MemStorage.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
