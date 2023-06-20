package storage

import (
	"context"
	"os"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/logger"
)

func TestState_SaveLoad(t *testing.T) {

	ctx := context.Background()

	ts := newMemStorage()
	ts.SetFloat64Value(ctx, "test1", 1.2)
	ts.AddInt64Value(ctx, "test4", 5)

	l, err := logger.NewLogger()
	if err != nil {
		t.Errorf("TestState_SaveLoad err: %v", err)
	}

	type fields struct {
		path          string
		storeInterval int
		stg           *MemStorage
		logger        *logger.AppLogger
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
				logger:        l,
			},
			wantErr: false,
		},
		{
			name: "#2 case",
			fields: fields{
				path:          newTempFile(t),
				storeInterval: 10,
				stg:           ts,
				logger:        l,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			defer os.Remove(tt.fields.path)

			st := &Filestorage{
				path:          tt.fields.path,
				storeInterval: tt.fields.storeInterval,
				MemStorage:    tt.fields.stg,
				logger:        tt.fields.logger.SugaredLogger,
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

	td := os.TempDir()

	f, err := os.CreateTemp(td, "*")
	if err != nil {
		t.Errorf("cannot create new temp file for filestorage tests: %v", err)
	}
	defer f.Close()

	return f.Name()

}
