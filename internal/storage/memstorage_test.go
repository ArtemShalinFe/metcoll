package storage

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {

	want := &MemStorage{
		mutex:       &sync.Mutex{},
		dataInt64:   make(map[string]int64),
		dataFloat64: make(map[string]float64),
	}

	t.Run("Test mem storage constructor", func(t *testing.T) {
		if got := newMemStorage(); !reflect.DeepEqual(got, want) {
			t.Errorf("NewMemStorage() = %v, want %v", got, want)
		}
	})
}

func TestMemStorage_GetFloat64Value(t *testing.T) {

	ctx := context.Background()

	ts := newMemStorage()
	ts.SetFloat64Value(ctx, "test1", 1.0)
	ts.SetFloat64Value(ctx, "test2", 2.0)
	ts.SetFloat64Value(ctx, "test3", 4.0)

	type args struct {
		Key string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr error
	}{
		{name: "test get storage value 1 - positive case",
			args:    args{Key: "test1"},
			want:    1.0,
			wantErr: nil,
		},
		{name: "test get storage value 2 - positive case",
			args:    args{Key: "test2"},
			want:    2.0,
			wantErr: nil,
		},
		{name: "test get storage value 4 - positive case",
			args:    args{Key: "test3"},
			want:    4.0,
			wantErr: nil,
		},
		{name: "test get storage value none - negative case",
			args:    args{Key: "test4"},
			want:    0.0,
			wantErr: ErrNoRows,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			value, err := ts.GetFloat64Value(ctx, tt.args.Key)
			if value != tt.want {
				t.Errorf("MemStorage.GetFloat64Value() = %v, want %v", value, tt.want)
			}
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestMemStorage_GetInt64Value(t *testing.T) {

	ctx := context.Background()

	ts := newMemStorage()
	ts.AddInt64Value(ctx, "test1", 1)
	ts.AddInt64Value(ctx, "test2", 2)
	ts.AddInt64Value(ctx, "test3", 3)
	ts.AddInt64Value(ctx, "test3", 3)

	type args struct {
		Key string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr error
	}{
		{name: "test get storage value 1 - positive case",
			args:    args{Key: "test1"},
			want:    1,
			wantErr: nil,
		},
		{name: "test get storage value 2 - positive case",
			args:    args{Key: "test2"},
			want:    2,
			wantErr: nil,
		},
		{name: "test get storage value 4 - negative case",
			args:    args{Key: "test3"},
			want:    6,
			wantErr: nil,
		},
		{name: "test get storage value 0 - negative case",
			args:    args{Key: "test9"},
			want:    0,
			wantErr: ErrNoRows,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			value, err := ts.GetInt64Value(ctx, tt.args.Key)
			if value != tt.want {
				t.Errorf("MemStorage.GetInt64Value() = %v, want %v", value, tt.want)
			}
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestMemStorage_GetDataList(t *testing.T) {

	ctx := context.Background()

	ts := newMemStorage()
	ts.SetFloat64Value(ctx, "test1", 1.2)
	ts.AddInt64Value(ctx, "test4", 5)

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

			got, err := tt.ms.GetDataList(ctx)
			if err != nil {
				t.Errorf("TestMemStorage_GetDataList err: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.GetDataList() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestGetSetState(t *testing.T) {

	ctx := context.Background()

	ts := newMemStorage()
	ts.SetFloat64Value(ctx, "test1", 1.2)
	ts.AddInt64Value(ctx, "test4", 5)

	tsb := ts
	b, err := ts.GetState()
	if err != nil {
		t.Error(err)
	}

	if err := ts.SetState(b); err != nil {
		t.Error(err)
	}

	assert.Equal(t, tsb, ts)

}
