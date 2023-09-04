package configuration

import (
	"testing"
)

func TestConfig_String(t *testing.T) {
	type fields struct {
		Address         string
		FileStoragePath string
		Database        string
		Key             []byte
		StoreInterval   int
		Restore         bool
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "check print hashkey",
			fields: fields{
				Address:         "nope",
				Key:             []byte("testKey"),
				FileStoragePath: "test",
				Database:        "somedsn",
				StoreInterval:   1,
				Restore:         false,
			},
			want: "Addres: nope, StoreInterval: 1, Restore: false, DSN: somedsn, FS path: test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Address:         tt.fields.Address,
				FileStoragePath: tt.fields.FileStoragePath,
				Database:        tt.fields.Database,
				Key:             tt.fields.Key,
				StoreInterval:   tt.fields.StoreInterval,
				Restore:         tt.fields.Restore,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("Config.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
