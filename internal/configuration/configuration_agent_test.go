package configuration

import (
	"flag"
	"fmt"
	"os"
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
