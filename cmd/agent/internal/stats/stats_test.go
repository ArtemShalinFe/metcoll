package stats

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/go-playground/assert"
	"github.com/stretchr/testify/require"
)

func TestStats_Update(t *testing.T) {
	type fields struct {
		memStats    *runtime.MemStats
		PollCount   int64
		RandomValue int64
	}
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
				RandomValue: time.Now().Unix(),
			},
			wantPollCount: 1,
		},
		{
			name: "negative case",
			fields: fields{
				memStats:    &runtime.MemStats{},
				PollCount:   987654141,
				RandomValue: time.Now().Unix(),
			},
			wantPollCount: 987654142,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{
				memStats:    tt.fields.memStats,
				PollCount:   tt.fields.PollCount,
				RandomValue: tt.fields.RandomValue,
			}
			s.Update()
			assert.Equal(t, tt.wantPollCount, s.PollCount)
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

		rGm := requiredGaugeMetrics()
		rCm := requiredCounterMetrics()

		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{
				memStats:    tt.fields.memStats,
				PollCount:   tt.fields.PollCount,
				RandomValue: tt.fields.RandomValue,
			}
			s.Update()

			gaugeData := s.GetReportData()[gauge]
			counterData := s.GetReportData()[counter]

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

func requiredGaugeMetrics() [28]string {

	return [28]string{

		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue"}

}

func requiredCounterMetrics() [1]string {

	return [1]string{
		"PollCount",
	}

}
