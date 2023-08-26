package configuration

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckConfigPriority(t *testing.T) {
	var c ConfigAgent

	const a = "localhost:8090"

	serverFS := flag.NewFlagSet("a", flag.ContinueOnError)
	serverFS.StringVar(&c.Server, "a", a, "server end point")

	limitFS := flag.NewFlagSet("l", flag.ContinueOnError)
	limitFS.IntVar(&c.Limit, "l", 1, "limit")

	if err := os.Setenv("ADDRESS", a); err != nil {
		fmt.Printf("set env ADDRESS err: %v", err)
		return
	}

	defer func() {
		if err := os.Unsetenv("ADDRESS"); err != nil {
			fmt.Printf("set env ADDRESS err: %v", err)
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
		fields fields
		want   string
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
			want: "Addres: nope, ReportInterval: 1, PollInterval: 1, Limit: 1",
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

func TestParseAgent(t *testing.T) {
	tests := []struct {
		name    string
		want    *ConfigAgent
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAgent()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAgent() = %v, want %v", got, tt.want)
			}
		})
	}
}
