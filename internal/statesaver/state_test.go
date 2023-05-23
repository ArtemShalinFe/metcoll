package statesaver

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

type testStorage struct {
	name string
}

func (s *testStorage) GetState() ([]byte, error) {
	return json.Marshal(&s)
}

func (s *testStorage) SetState(data []byte) error {
	return json.Unmarshal(data, &s)
}

type testLogger struct{}

func (tl *testLogger) Info(args ...any) {
	log.Println(args...)
}

func (tl *testLogger) Error(args ...any) {
	log.Println(args...)
}

func TestState_SaveLoad(t *testing.T) {
	type fields struct {
		fileStoragePath string
		storeInterval   int
		stg             StorageState
		logger          Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "#1 case",
			fields: fields{
				fileStoragePath: "/tmp/tests-state-save-storage-1.json",
				storeInterval:   10,
				stg:             &testStorage{name: "test1"},
				logger:          &testLogger{},
			},
			wantErr: false,
		},
		{
			name: "#2 case",
			fields: fields{
				fileStoragePath: "/tmp/tests-state-save-storage-2.json",
				storeInterval:   10,
				stg:             &testStorage{name: "test2"},
				logger:          &testLogger{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			defer os.Remove(tt.fields.fileStoragePath)

			st := &State{
				fileStoragePath: tt.fields.fileStoragePath,
				storeInterval:   tt.fields.storeInterval,
				stg:             tt.fields.stg,
				logger:          tt.fields.logger,
			}
			if err := st.Save(); (err != nil) != tt.wantErr {
				t.Errorf("State.Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := st.Load(); (err != nil) != tt.wantErr {
				t.Errorf("State.Load() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}
