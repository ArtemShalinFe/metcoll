package storage

import (
	"context"
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestState_SaveLoad(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, "test12", 1.2); err != nil {
		t.Error(err)
	}
	if _, err := ts.AddInt64Value(ctx, "test5", 5); err != nil {
		t.Error(err)
	}

	zl, err := zap.NewProduction()
	if err != nil {
		t.Errorf("cannot init zap-logger err: %v ", err)
	}

	sl := zl.Sugar()

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
