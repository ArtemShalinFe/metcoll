package main

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

var cfg *configuration.ConfigAgent

type mockClient struct{}

func (c *mockClient) BatchUpdate(ctx context.Context, m []*metrics.Metrics) error {
	return nil
}

func TestMain(m *testing.M) {
	c, err := configuration.ParseAgent()
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
			d := time.Duration(tt.args.reportInterval) * time.Second
			if got := isTimeToPushReport(tt.args.lastReportPush, tt.args.now, d); got != tt.want {
				t.Errorf("isTimeToPushReport() = %v, want %v", got, tt.want)
			}
		})
	}

}

func Test_pushReport(t *testing.T) {

	ctx := context.Background()

	type args struct {
		conn metcollClient
		s    *stats.Stats
		cfg  *configuration.ConfigAgent
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
			if err := pushReport(ctx, tt.args.conn, tt.args.s, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("pushReport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
