//go:build usetempdir
// +build usetempdir

package configuration

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
				Key:             []byte("testHashKey"),
				FileStoragePath: "test",
				Database:        "somedsn",
				StoreInterval:   1,
				Restore:         false,
			},
			want: "Addres: nope, StoreInterval: 1, Restore: false, DSN: somedsn, FS path: test, Path: ",
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

func Test_readConfigFromENV(t *testing.T) {
	const envAddressName = "ADDRESS"

	t.Setenv(envAddressName, localhost8090)
	t.Setenv(envHashKey, envAddressName)

	want := newConfig()
	want.Address = localhost8090

	tests := []struct {
		want    *Config
		name    string
		wantErr bool
	}{
		{
			name:    "#1",
			want:    want,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := readConfigFromENV()
			if (err != nil) != tt.wantErr {
				t.Errorf("readFromENV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got.Address, tt.want.Address) {
				t.Errorf("readFromENV() = %v, want %v", got.Address, tt.want.Address)
			}
		})
	}
}

func Test_readConfigFromFile(t *testing.T) {
	jsonConfig := newConfigFile(t,
		`{
			"address": "localhost:8080",
			"restore": true,
			"store_interval": "1s",
			"store_file": "/tmp/metrics-db.json", 
			"database_dsn": "", 
			"crypto_key": "/path/to/key.pem"
		}`)

	jsonConfig2 := newConfigFile(t,
		`{
			"address": "localhost:8090",
			"restore": true,
			"store_interval": "1m", 
			"store_file": "/tmp/metrics-db.json", 
			"database_dsn": "", 
			"crypto_key": "/path/to/key.pem"
		}`)

	jsonConfigErr := newConfigFile(t,
		`{
		"store_interval": "1masdasd",
	}`)

	want := newConfig()
	want.StoreInterval = 1

	want2 := newConfig()
	want2.Address = "localhost:8090"
	want2.StoreInterval = 60

	wantErr := newConfig()

	tests := []struct {
		want    *Config
		name    string
		path    string
		wantErr bool
	}{
		{
			name: "#1",
			path: jsonConfig,
			want: want,
		},
		{
			name: "#2",
			path: jsonConfig2,
			want: want2,
		},
		{
			name:    "#3",
			path:    jsonConfigErr,
			want:    wantErr,
			wantErr: true,
		},
		{
			name:    "#4",
			path:    "",
			want:    newConfig(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := readConfigFromFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_setFromConfigs(t *testing.T) {
	clc := newConfig()
	clc.Database = "changed"

	envc := newConfig()
	envc.Database = "changed2"

	fc := newConfig()
	fc.Database = "changed3"
	fc.Restore = false

	want := &Config{
		Address:         defaultMetcollAddress,
		FileStoragePath: defaultFileStoragePath,
		Database:        envc.Database,
		Key:             []byte(defaultHashKey),
		StoreInterval:   defaultStoreInterval,
		Restore:         fc.Restore,
		Path:            "",
	}

	type args struct {
		configCL   *Config
		configENV  *Config
		configFile *Config
		path       string
	}
	tests := []struct {
		name string
		want *Config
		args args
	}{
		{
			name: "#1",
			want: want,
			args: args{
				configCL:   clc,
				configENV:  envc,
				configFile: fc,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := newConfig()
			c.setFromConfigs(tt.args.configCL, tt.args.configENV, tt.args.configFile, tt.args.path)

			assert.Equal(t, tt.want.Database, c.Database, "DSN path comparing")
			assert.Equal(t, tt.want.Restore, c.Restore, "Restore storage comparing")
		})
	}
}

func newConfigFile(t *testing.T, jsonText string) string {
	t.Helper()
	td := os.TempDir()

	f, err := os.CreateTemp(td, "*.json")
	if err != nil {
		t.Errorf("cannot create new json file for config tests: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("cannot close json file for config tests: %v", err)
		}
	}()

	if _, err := f.Write([]byte(jsonText)); err != nil {
		t.Errorf("cannot write json text in file for config tests: %v", err)
	}

	return f.Name()
}
