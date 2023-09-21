//go:build usetempdir
// +build usetempdir

package configuration

import (
	"flag"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	localhost8090  = "localhost:8090"
	envAddressName = "ADDRESS"
)

func TestCheckConfigAgentPriority(t *testing.T) {
	var c ConfigAgent

	serverFS := flag.NewFlagSet(metcollAddressFlagName, flag.ContinueOnError)
	serverFS.StringVar(&c.Server, metcollAddressFlagName, localhost8090, "metcollserver  end point")

	limitFS := flag.NewFlagSet(limitFlagName, flag.ContinueOnError)
	limitFS.IntVar(&c.Limit, limitFlagName, 1, "limit")

	t.Setenv(envAddressName, localhost8090)

	flag.Parse()

	tests := []struct {
		want    *ConfigAgent
		name    string
		wantErr bool
	}{
		{
			name:    "Check priority",
			want:    &c,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgent()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.Server, got.Server)
			assert.Equal(t, tt.want.Limit, got.Limit)
		})
	}
}

func TestConfigAgent_String(t *testing.T) {
	type fields struct {
		Server         string
		Key            []byte
		PollInterval   int
		ReportInterval int
		Limit          int
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "check print hashkey",
			fields: fields{
				Server:         "nope",
				Key:            []byte("testKey"),
				PollInterval:   1,
				ReportInterval: 1,
				Limit:          1,
			},
			want: "Addres: nope, ReportInterval: 1, PollInterval: 1, Limit: 1, Path: ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigAgent{
				Server:         tt.fields.Server,
				Key:            tt.fields.Key,
				PollInterval:   tt.fields.PollInterval,
				ReportInterval: tt.fields.ReportInterval,
				Limit:          tt.fields.Limit,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("ConfigAgent.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readConfigAgentFromENV(t *testing.T) {
	const a = "localhost:8090"
	const envAddressName = "ADDRESS"

	t.Setenv(envAddressName, a)

	want := newConfigAgent()
	want.Server = a

	tests := []struct {
		want    *ConfigAgent
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
			got, err := readConfigAgentFromENV()
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfigAgentFromENV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got.Server, tt.want.Server) {
				t.Errorf("readConfigAgentFromENV() = %v, want %v", got.Server, tt.want.Server)
			}
		})
	}
}

func Test_readConfigAgentFromFile(t *testing.T) {
	jsonConfig := newAgentConfigFile(t,
		`{
		"address": "localhost:8080",
		"report_interval": "1s",
		"poll_interval": "1s",
		"crypto_key": "/path/to/key.pem"
	}`)

	jsonConfig2 := newAgentConfigFile(t,
		`{
		"address": "localhost:8090",
		"report_interval": "1m",
		"poll_interval": "1h",
		"crypto_key": "/path/to/key.pem"
	}`)

	jsonConfigErr := newAgentConfigFile(t,
		`{
		"report_interval": "1masdasd",
	}`)

	want := newConfigAgent()
	want.ReportInterval = 1
	want.PollInterval = 1

	reportInterval := 60
	pollInterval := 3600

	want2 := newConfigAgent()
	want2.Server = localhost8090
	want2.ReportInterval = reportInterval
	want2.PollInterval = pollInterval

	wantErr := newConfigAgent()

	tests := []struct {
		want    *ConfigAgent
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := readConfigAgentFromFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfigAgentFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfigAgentFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newAgentConfigFile(t *testing.T, jsonText string) string {
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
