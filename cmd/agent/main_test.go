package main

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

var cfg *Config

type mockClient struct{}

func (m *mockClient) Push(mType string, Name string, Value string) error {
	return nil
}

func TestMain(m *testing.M) {
	c, err := parseConfig()
	if err != nil {
		log.Print(err)
	}
	cfg = c

	os.Exit(m.Run())

}

func Test_isTimeToPushReport(t *testing.T) {

	now := time.Now()

	type args struct {
		lastReportPush time.Time
		now            time.Time
		reportInterval int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive case",
			args: args{
				lastReportPush: now,
				now:            time.Now(),
				reportInterval: cfg.ReportInterval},
			want: false,
		},
		{
			name: "positive case",
			args: args{
				lastReportPush: now.Add(time.Second * time.Duration(cfg.ReportInterval) * -1),
				now:            time.Now(),
				reportInterval: cfg.ReportInterval},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimeToPushReport(tt.args.lastReportPush, tt.args.now, tt.args.reportInterval); got != tt.want {
				t.Errorf("isTimeToPushReport() = %v, want %v", got, tt.want)
			}
		})
	}

}

func Test_pushReport(t *testing.T) {

	type args struct {
		conn client
		s    *stats.Stats
		cfg  *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				conn: &mockClient{},
				s:    stats.NewStats(),
				cfg:  cfg,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := pushReport(tt.args.conn, tt.args.s, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("pushReport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
