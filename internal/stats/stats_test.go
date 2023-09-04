package stats

import (
	"context"
	"testing"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/go-playground/assert"
)

func BenchmarkGetReportData(b *testing.B) {
	b.StopTimer()

	ctx := context.Background()

	conf := &configuration.ConfigAgent{}
	conf.Key = []byte("")
	conf.Limit = 1
	conf.PollInterval = 1
	conf.ReportInterval = 2

	mcs := make(chan []*metrics.Metrics, conf.Limit)

	s := NewStats()
	s.RunCollectBatchStats(ctx, conf, mcs)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.GetReportData(ctx)
	}
}

func TestStats_RunCollectBatchStats(t *testing.T) {
	ctx := context.Background()

	conf := &configuration.ConfigAgent{}
	conf.Key = []byte("")
	conf.Limit = 1
	conf.PollInterval = 0
	conf.ReportInterval = 0

	mcs := make(chan []*metrics.Metrics, conf.Limit)

	s := NewStats()

	type args struct {
		cfg *configuration.ConfigAgent
		ms  chan []*metrics.Metrics
	}
	tests := []struct {
		args  args
		stats *Stats
		name  string
	}{
		{
			name:  "check collect batch stats",
			stats: s,
			args: args{
				cfg: conf,
				ms:  mcs,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctxWithTimeout, cancelCtx := context.WithTimeout(ctx, time.Second*5)
			defer cancelCtx()
			tt.stats.RunCollectBatchStats(ctxWithTimeout, tt.args.cfg, tt.args.ms)
			select {
			case <-ctx.Done():
				return
			case <-tt.args.ms:
			}
		})
	}
}

func TestStats_ClearPollCount(t *testing.T) {
	s := NewStats()
	s.pollCount = 10

	tests := []struct {
		stats *Stats
		name  string
	}{
		{
			name:  "cleanup",
			stats: s,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.stats.ClearPollCount()

			assert.Equal(t, tt.stats.pollCount, int64(0))
		})
	}
}
