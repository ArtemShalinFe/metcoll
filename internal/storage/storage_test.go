// The package describes the interaction of the server with various sources of metrics storage.
// Metrics can be stored in memory, in a file on disk, in a database.
package storage

import (
	"context"
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

	fs, err := newFilestorage(ts, zap.L().Sugar(), newTempFile(t), 0, false)
	if err != nil {
		t.Error(err)
	}

	type args struct {
		ctx context.Context
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
				ctx: ctx,
				cfg: &configuration.Config{
					FileStoragePath: newTempFile(t),
				},
			},
			want:    fs,
			wantErr: false,
		},
		{
			name: "#2 case sqlstorage",
			args: args{
				ctx: ctx,
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
			got, err := InitStorage(tt.args.ctx, tt.args.cfg, tt.args.l)
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
