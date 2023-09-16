package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-playground/assert"
	"go.uber.org/zap"
)

const test5 = "test5"
const test12 = "test12"

func TestState_SaveLoad(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, test12, 1.2); err != nil {
		t.Error(err)
	}
	if _, err := ts.AddInt64Value(ctx, test5, 5); err != nil {
		t.Error(err)
	}

	sl := zap.L().Sugar()

	type fields struct {
		stg           *MemStorage
		logger        *zap.SugaredLogger
		path          string
		storeInterval int
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "#1 case",
			fields: fields{
				path:          newTempFile(t),
				storeInterval: 10,
				stg:           ts,
				logger:        sl,
			},
			wantErr: false,
		},
		{
			name: "#2 case",
			fields: fields{
				path:          newTempFile(t),
				storeInterval: 10,
				stg:           ts,
				logger:        sl,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := os.Remove(tt.fields.path); err != nil {
					t.Errorf("cannot remove temp file for filestorage tests: %v", err)
				}
			}()

			st := &Filestorage{
				path:          tt.fields.path,
				storeInterval: tt.fields.storeInterval,
				MemStorage:    tt.fields.stg,
				logger:        tt.fields.logger,
			}
			if err := st.Save(st.MemStorage); (err != nil) != tt.wantErr {
				t.Errorf("State.Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := st.Load(st.MemStorage); (err != nil) != tt.wantErr {
				t.Errorf("State.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func newTempFile(t *testing.T) string {
	t.Helper()
	td := os.TempDir()

	f, err := os.CreateTemp(td, "*")
	if err != nil {
		t.Errorf("cannot create new temp file for filestorage tests: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("cannot close temp file for filestorage tests: %v", err)
		}
	}()

	return f.Name()
}

func TestFilestorage_Interrupt(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, test12, 1.2); err != nil {
		t.Error(err)
	}
	if _, err := ts.AddInt64Value(ctx, test5, 5); err != nil {
		t.Error(err)
	}

	sl := zap.L().Sugar()

	type fields struct {
		MemStorage    *MemStorage
		logger        *zap.SugaredLogger
		path          string
		storeInterval int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "#1 interrupt test",
			fields: fields{
				MemStorage:    ts,
				logger:        sl,
				path:          newTempFile(t),
				storeInterval: 10,
			},
			wantErr: false,
		},
		{
			name: "#2 interrupt test",
			fields: fields{
				MemStorage:    ts,
				logger:        sl,
				path:          "",
				storeInterval: 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &Filestorage{
				MemStorage:    tt.fields.MemStorage,
				logger:        tt.fields.logger,
				path:          tt.fields.path,
				storeInterval: tt.fields.storeInterval,
			}
			if err := fs.Interrupt(); (err != nil) != tt.wantErr {
				t.Errorf("Filestorage.Interrupt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFilestorage_SetFloat64Value(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, test12, 1.2); err != nil {
		t.Error(err)
	}

	fs, err := newFilestorage(ctx, ts, zap.L().Sugar(), newTempFile(t), 0, false)
	if err != nil {
		t.Error(err)
	}

	type args struct {
		key   string
		value float64
	}
	tests := []struct {
		name    string
		fs      *Filestorage
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "#1 positive case",
			fs:   fs,
			args: args{
				key:   test12,
				value: 1.3,
			},
			want:    1.3,
			wantErr: false,
		},
		{
			name: "#2 positive case",
			fs:   fs,
			args: args{
				key:   test5,
				value: 6,
			},
			want:    6,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.SetFloat64Value(ctx, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filestorage.SetFloat64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Filestorage.SetFloat64Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilestorage_AddInt64Value(t *testing.T) {
	ctx := context.Background()

	const test1 = "test1"

	ts := newMemStorage()
	if _, err := ts.AddInt64Value(ctx, test1, 1); err != nil {
		t.Error(err)
	}

	fs, err := newFilestorage(ctx, ts, zap.L().Sugar(), newTempFile(t), 0, false)
	if err != nil {
		t.Error(err)
	}

	type args struct {
		key   string
		value int64
	}
	tests := []struct {
		name    string
		fs      *Filestorage
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "#1 positive case",
			fs:   fs,
			args: args{
				key:   test1,
				value: 1,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "#2 positive case",
			fs:   fs,
			args: args{

				key:   "test6",
				value: 6,
			},
			want:    6,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := fs.AddInt64Value(ctx, tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filestorage.AddInt64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Filestorage.AddInt64Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilestorage_BatchAddInt64Value(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.AddInt64Value(ctx, "test7", 7); err != nil {
		t.Error(err)
	}

	fs, err := newFilestorage(ctx, ts, zap.L().Sugar(), newTempFile(t), 0, false)
	if err != nil {
		t.Error(err)
	}

	c := make(map[string]int64)
	c["test7"] = 7
	c["test8"] = 8

	wantC := make(map[string]int64)
	wantC["test7"] = 14
	wantC["test8"] = 8

	type args struct {
		counters map[string]int64
	}
	tests := []struct {
		args    args
		fs      *Filestorage
		want    map[string]int64
		name    string
		wantErr bool
	}{
		{
			name: "#1 case",
			fs:   fs,
			args: args{
				counters: c,
			},
			want:    wantC,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := tt.fs.BatchAddInt64Value(ctx, tt.args.counters)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filestorage.BatchAddInt64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})

		for k, v := range wantC {
			value, err := tt.fs.GetInt64Value(ctx, k)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, value, v)
		}
	}
}

func TestFilestorage_BatchSetFloat64Value(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, "test123", 1.3); err != nil {
		t.Error(err)
	}

	fs, err := newFilestorage(ctx, ts, zap.L().Sugar(), newTempFile(t), 0, false)
	if err != nil {
		t.Error(err)
	}

	gauges := make(map[string]float64)
	gauges["test123"] = 1.4
	gauges["test14"] = 0.7

	type args struct {
		gauges map[string]float64
	}
	tests := []struct {
		args    args
		fs      *Filestorage
		want    map[string]float64
		name    string
		wantErr bool
	}{
		{
			name: "#1 case",
			fs:   fs,
			args: args{
				gauges: gauges,
			},
			want:    gauges,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := tt.fs.BatchSetFloat64Value(ctx, tt.args.gauges)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filestorage.BatchSetFloat64Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})

		for k, v := range gauges {
			value, err := tt.fs.GetFloat64Value(ctx, k)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, value, v)
		}
	}
}

func TestFilestorage_runIntervalStateSaving(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, "test23", 2.3); err != nil {
		t.Error(err)
	}

	type fields struct {
		MemStorage    *MemStorage
		logger        *zap.SugaredLogger
		path          string
		storeInterval int
		sleepTime     int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test interval saving",
			fields: fields{
				MemStorage:    ts,
				logger:        zap.L().Sugar(),
				path:          newTempFile(t),
				storeInterval: 1,
				sleepTime:     3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &Filestorage{
				MemStorage:    tt.fields.MemStorage,
				logger:        tt.fields.logger,
				path:          tt.fields.path,
				storeInterval: tt.fields.storeInterval,
			}

			tctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(tt.fields.sleepTime))
			defer cancel()

			fs.runIntervalStateSaving(tctx)

			time.Sleep(time.Duration(tt.fields.sleepTime) * time.Second)

			f, err := os.Open(tt.fields.path)
			if err != nil {
				t.Errorf("an erorr occured when opening file, err: %v", err)
				return
			}

			i, err := f.Stat()
			if err != nil {
				t.Errorf("an erorr occured when getting file stats, err: %v", err)
				return
			}

			assert.NotEqual(t, i.Size(), 0)
		})
	}
}

func TestFilestorage_Ping(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, "test71", 7.1); err != nil {
		t.Error(err)
	}

	type fields struct {
		MemStorage    *MemStorage
		logger        *zap.SugaredLogger
		path          string
		storeInterval int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "#1",
			fields: fields{
				MemStorage:    ts,
				logger:        zap.L().Sugar(),
				path:          newTempFile(t),
				storeInterval: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &Filestorage{
				MemStorage:    tt.fields.MemStorage,
				logger:        tt.fields.logger,
				path:          tt.fields.path,
				storeInterval: tt.fields.storeInterval,
			}
			if err := fs.Ping(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Filestorage.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
