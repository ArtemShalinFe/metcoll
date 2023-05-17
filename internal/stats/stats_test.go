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
		rGm := requiredGaugeMetrics()
		rCm := requiredCounterMetrics()

		t.Run(tt.name, func(t *testing.T) {
			s := &Stats{
				memStats:    tt.fields.memStats,
				pollCount:   tt.fields.PollCount,
				randomValue: tt.fields.RandomValue,
			}
			s.Update()

			gaugeData := s.GetReportData()[gaugeMetric]
			counterData := s.GetReportData()[counterMetric]

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

func requiredGaugeMetrics() []string {

	var rGm []string
	rGm = append(rGm, "Alloc")
	rGm = append(rGm, "BuckHashSys")
	rGm = append(rGm, "Frees")
	rGm = append(rGm, "GCCPUFraction")
	rGm = append(rGm, "GCSys")
	rGm = append(rGm, "HeapAlloc")
	rGm = append(rGm, "HeapIdle")
	rGm = append(rGm, "HeapInuse")
	rGm = append(rGm, "HeapObjects")
	rGm = append(rGm, "HeapReleased")
	rGm = append(rGm, "HeapSys")
	rGm = append(rGm, "LastGC")
	rGm = append(rGm, "Lookups")
	rGm = append(rGm, "MCacheInuse")
	rGm = append(rGm, "MCacheSys")
	rGm = append(rGm, "MSpanInuse")
	rGm = append(rGm, "MSpanSys")
	rGm = append(rGm, "Mallocs")
	rGm = append(rGm, "NextGC")
	rGm = append(rGm, "NumForcedGC")
	rGm = append(rGm, "NumGC")
	rGm = append(rGm, "OtherSys")
	rGm = append(rGm, "PauseTotalNs")
	rGm = append(rGm, "StackInuse")
	rGm = append(rGm, "StackSys")
	rGm = append(rGm, "Sys")
	rGm = append(rGm, "TotalAlloc")
	rGm = append(rGm, "RandomValue")

	return rGm

}

func requiredCounterMetrics() []string {

	var rCm []string
	rCm = append(rCm, "PollCount")

	return rCm

}

func TestIsPollCountMetric(t *testing.T) {

	type args struct {
		metType string
		name    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "positive case #1",
			args: args{
				metType: counterMetric,
				name:    pollCount,
			},
			want: true,
		},
		{
			name: "negative case #1",
			args: args{
				metType: gaugeMetric,
				name:    pollCount,
			},
			want: false,
		},
		{
			name: "negative case #2",
			args: args{
				metType: gaugeMetric,
				name:    "TotalAlloc",
			},
			want: false,
		},
		{
			name: "negative case #3",
			args: args{
				metType: counterMetric,
				name:    "TotalAlloc",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPollCountMetric(tt.args.metType, tt.args.name); got != tt.want {
				t.Errorf("IsPollCountMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}
