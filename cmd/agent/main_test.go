package main

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	mock_main "github.com/ArtemShalinFe/metcoll/cmd/agent/mock"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

func Test_isTimeToPushReport(t *testing.T) {

	now := time.Now()

	cfg, err := configuration.ParseAgent()
	if err != nil {
		t.Errorf("Test_isTimeToPushReport err %v", err)
	}

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

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockClient := mock_main.NewMockmetcollClient(ctl)

	var ms []*metrics.Metrics
	for _, data := range stats.NewStats().GetReportData(ctx) {
		for _, metric := range data {
			ms = append(ms, metric)
		}
	}

	gomock.InOrder(
		mockClient.EXPECT().BatchUpdate(ctx, ms).Return(nil),
	)

	type args struct {
		conn metcollClient
		ms   []*metrics.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				conn: mockClient,
				ms:   ms,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := pushReport(ctx, tt.args.conn, tt.args.ms); (err != nil) != tt.wantErr {
				t.Errorf("pushReport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
