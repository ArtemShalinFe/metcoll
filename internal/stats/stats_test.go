package stats

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/go-playground/assert"
	"github.com/stretchr/testify/require"
)

func TestStats_Update(t *testing.T) {

	type fields struct {
		memStats    *runtime.MemStats
		PollCount   int64
		RandomValue int64
	}

	now := time.Now()

	tests := []struct {
		name          string
		fields        fields
		wantPollCount int64
	}{
		{
			name: "positive case",
			fields: fields{
				memStats:    &runtime.MemStats{},
				PollCount:   0,
				RandomValue: now.Unix(),
			},
			wantPollCount: 1,
		},
	}

	time.Sleep(1 * time.Second)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := NewStats()
			s.Update()

			assert.Equal(t, tt.wantPollCount, s.pollCount)
			assert.NotEqual(t, tt.fields.RandomValue, s.randomValue)

			s.ClearPollCount()
			assert.Equal(t, int64(0), s.pollCount)

		})
	}
}

func TestStats_GetReportData(t *testing.T) {
	type fields struct {
		memStats    *runtime.MemStats
		PollCount   int64
		RandomValue int64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "positive case",
			fields: fields{
				memStats:    &runtime.MemStats{},
				PollCount:   0,
				RandomValue: time.Now().Unix(),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		rGm := gaugeMetrics()
		rCm := counterMetrics()

		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{
				memStats:    tt.fields.memStats,
				pollCount:   tt.fields.PollCount,
				randomValue: tt.fields.RandomValue,
			}
			s.Update()

			gaugeData := s.GetReportData()[metrics.GaugeMetric]
			counterData := s.GetReportData()[metrics.CounterMetric]

			for name := range gaugeData {
				require.Contains(t, rGm, name, fmt.Sprintf("Tests GetReportData gaugeData not contain required gauge metrics %s", name))
				require.NotContains(t, rCm, name, fmt.Sprintf("Tests GetReportData gaugeData contain required counter metrics %s", name))
			}

			for name := range counterData {
				require.Contains(t, rCm, name, fmt.Sprintf("Tests GetReportData counterData not contain required counter metrics %s", name))
				require.NotContains(t, rGm, name, fmt.Sprintf("Tests GetReportData counterData contain required gauge metrics %s", name))
			}

		})
	}
}
