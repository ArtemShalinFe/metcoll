package filestorage

import (
	"log"
	"os"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

type testLogger struct{}

func (tl *testLogger) Info(args ...any) {
	log.Println(args...)
}

func (tl *testLogger) Error(args ...any) {
	log.Println(args...)
}

func TestState_SaveLoad(t *testing.T) {

	ts := storage.NewMemStorage()
	ts.SetFloat64Value("test1", 1.2)
	ts.AddInt64Value("test4", 5)

	type fields struct {
		path          string
		storeInterval int
		stg           *storage.MemStorage
		logger        Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "#1 case",
			fields: fields{
				path:          "/tmp/tests-state-save-storage-1.json",
				storeInterval: 10,
				stg:           ts,
				logger:        &testLogger{},
			},
			wantErr: false,
		},
		{
			name: "#2 case",
			fields: fields{
				path:          "/tmp/tests-state-save-storage-2.json",
				storeInterval: 10,
				stg:           ts,
				logger:        &testLogger{},
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
