package storage

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	want := &MemStorage{
		Data: make(map[string]interface{}),
	}

	t.Run("Test mem storage constructor", func(t *testing.T) {
		if got := NewMemStorage(); !reflect.DeepEqual(got, want) {
			t.Errorf("NewMemStorage() = %v, want %v", got, want)
		}
	})
}

func TestMemStorage_GetFloat64Value(t *testing.T) {
	ts := NewMemStorage()
	ts.SetFloat64Value("test1", 1.0)
	ts.SetFloat64Value("test2", 2.0)
	ts.SetFloat64Value("test3", 4.0)
	ts.SetInt64Value("test9", 9)

	type args struct {
		Key string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{name: "test get storage value 1 - positive case",
			args:    args{Key: "test1"},
			want:    1.0,
			wantErr: false,
		},
		{name: "test get storage value 2 - positive case",
			args:    args{Key: "test2"},
			want:    2.0,
			wantErr: false,
		},
		{name: "test get storage value 4 - positive case",
			args:    args{Key: "test3"},
			want:    4.0,
			wantErr: false,
		},
		{name: "test get storage value 0 - positive case",
			args:    args{Key: "test4"},
			want:    0,
			wantErr: false,
		},
		{name: "test get storage value 9 - negative case",
			args:    args{Key: "test9"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := ts.GetFloat64Value(tt.args.Key)
			if value != tt.want {
				t.Errorf("MemStorage.GetFloat64Value() = %v, want %v", value, tt.want)
			}

			if err != nil {
				assert.Equal(t, tt.wantErr, errors.Is(err, ErrWrongType))
			}
		})
	}
}

func TestMemStorage_GetInt64Value(t *testing.T) {
	ts := NewMemStorage()

	ts.SetInt64Value("test1", 1)
	ts.SetInt64Value("test2", 2)
	ts.SetInt64Value("test3", 3)
	ts.SetFloat64Value("test4", 3)

	type args struct {
		Key string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{name: "test get storage value 1 - positive case",
			args:    args{Key: "test1"},
			want:    1,
			wantErr: false,
		},
		{name: "test get storage value 2 - positive case",
			args:    args{Key: "test2"},
			want:    2,
			wantErr: false,
		},
		{name: "test get storage value 3 - positive case",
			args:    args{Key: "test3"},
			want:    3,
			wantErr: false,
		},
		{name: "test get storage value 4 - negative case",
			args:    args{Key: "test4"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := ts.GetInt64Value(tt.args.Key)
			if value != tt.want {
				t.Errorf("MemStorage.GetInt64Value() = %v, want %v", value, tt.want)
			}

			if err != nil {
				assert.Equal(t, tt.wantErr, errors.Is(err, ErrWrongType))
			}
		})
	}
}
