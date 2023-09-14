package configuration

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckConfigAgentPriority(t *testing.T) {
	var c ConfigAgent

	const a = "localhost:8090"
	const envAddressName = "ADDRESS"

	serverFS := flag.NewFlagSet("a", flag.ContinueOnError)
	serverFS.StringVar(&c.Server, "a", a, "metcollserver  end point")

	limitFS := flag.NewFlagSet("l", flag.ContinueOnError)
	limitFS.IntVar(&c.Limit, "l", 1, "limit")

	if err := os.Setenv(envAddressName, a); err != nil {
		fmt.Printf("set env ADDRESS err: %v", err)
		return
	}

	defer func() {
		if err := os.Unsetenv(envAddressName); err != nil {
			fmt.Printf("unset env ADDRESS err: %v", err)
		}
	}()

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

	if err := os.Setenv(envAddressName, a); err != nil {
		fmt.Printf("set env ADDRESS err: %v", err)
		return
	}

	defer func() {
		if err := os.Unsetenv(envAddressName); err != nil {
			fmt.Printf("unset env ADDRESS err: %v", err)
		}
	}()

	want := newConfigAgent()
	want.Server = a

	tests := []struct {
		name    string
		want    *ConfigAgent
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
				t.Errorf("readFromENV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, got.Server, tt.want.Server) {
				t.Errorf("readFromENV() = %v, want %v", got.Server, tt.want.Server)
			}
		})
	}
}

func Test_readConfigAgentFromFile(t *testing.T) {
	jsonConfig := newConfigAgentFile(t,
		`{
		"address": "localhost:8080",
		"report_interval": "1s",
		"poll_interval": "1s",
		"crypto_key": "/path/to/key.pem"
	}`)

	jsonConfig2 := newConfigAgentFile(t,
		`{
		"address": "localhost:8090",
		"report_interval": "1m",
		"poll_interval": "1h",
		"crypto_key": "/path/to/key.pem"
	}`)

	jsonConfigErr := newConfigAgentFile(t,
		`{
		"report_interval": "1masdasd",
	}`)

	want := newConfigAgent()
	want.ReportInterval = 1
	want.PollInterval = 1

	want2 := newConfigAgent()
	want2.Server = "localhost:8090"
	want2.ReportInterval = 60
	want2.PollInterval = 3600

	wantErr := newConfigAgent()

	tests := []struct {
		name    string
		path    string
		want    *ConfigAgent
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
				t.Errorf("readFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newConfigAgentFile(t *testing.T, jsonText string) string {
	t.Helper()
	td := os.TempDir()

	f, err := os.CreateTemp(td, "*.json")
	if err != nil {
		t.Errorf("cannot create new json file for config agent tests: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("cannot close json file for config agent tests: %v", err)
		}
	}()

	if _, err := f.Write([]byte(jsonText)); err != nil {
		t.Errorf("cannot write json text in file for config agent tests: %v", err)
	}

	return f.Name()
}
