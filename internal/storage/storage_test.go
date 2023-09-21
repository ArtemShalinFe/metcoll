//go:build usetempdir
// +build usetempdir

// The package describes the interaction of the server with various sources of metrics storage.
// Metrics can be stored in memory, in a file on disk, in a database.
package storage

import (
	"context"
	"os"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"go.uber.org/zap"
)

func TestInitStorage(t *testing.T) {
	ctx := context.Background()

	ts := newMemStorage()
	if _, err := ts.SetFloat64Value(ctx, "test12", 1.2); err != nil {
		t.Error(err)
	}

	fs, err := newFilestorage(ctx, ts, zap.L().Sugar(), newTempFile(t), 0, false)
	if err != nil {
		t.Error(err)
	}

	type args struct {
		cfg *configuration.Config
		l   *zap.SugaredLogger
	}
	tests := []struct {
		args    args
		want    Storage
		name    string
		wantErr bool
	}{
		{
			name: "#1 case filestorage",
			args: args{
				cfg: &configuration.Config{
					FileStoragePath: newFileStorageFile(t),
				},
			},
			want:    fs,
			wantErr: false,
		},
		{
			name: "#2 case sqlstorage",
			args: args{

				cfg: &configuration.Config{
					Database: "somestring",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitStorage(ctx, tt.args.cfg, tt.args.l)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch v := got.(type) {
			case nil:
			case *Filestorage:
			case *DB:
			case *MemStorage:
			default:
				t.Errorf("wrong type: %v", v)
			}
		})
	}
}

func newFileStorageFile(t *testing.T) string {
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
