package stats

import (
	"runtime"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type Stats struct {
	memStats    *runtime.MemStats
	pollCount   int64
	randomValue int64 // timestamp
}

func NewStats() *Stats {

	return &Stats{
		memStats:    &runtime.MemStats{},
		pollCount:   0,
		randomValue: time.Now().Unix(),
	}
}

func (s *Stats) Update() {

	runtime.ReadMemStats(s.memStats)
	s.randomValue = time.Now().Unix()
	s.pollCount = s.pollCount + 1

}

func (s *Stats) ClearPollCount() {
	s.pollCount = 0
}

func (s *Stats) GetReportData() map[string]map[string]*metrics.Metrics {

	gm := make(map[string]*metrics.Metrics)
	for _, id := range gaugeMetrics() {
		m := metrics.NewGaugeMetric(id, 0)
		m.Get(s)
		gm[id] = m
	}

	cm := make(map[string]*metrics.Metrics)
	for _, mid := range counterMetrics() {
		m := metrics.NewCounterMetric(mid, 0)
		m.Get(s)
		cm[mid] = m
	}

	reportData := make(map[string]map[string]*metrics.Metrics)
	reportData[metrics.GaugeMetric] = gm
	reportData[metrics.CounterMetric] = cm

	return reportData

}

func (s *Stats) GetFloat64Value(id string) (float64, bool) {

	switch id {
	case "Alloc":
		return float64(s.memStats.Alloc), true
	case "BuckHashSys":
		return float64(s.memStats.BuckHashSys), true
	case "GCCPUFraction":
		return float64(s.memStats.GCCPUFraction), true
	case "HeapAlloc":
		return float64(s.memStats.HeapAlloc), true
	case "GCSys":
		return float64(s.memStats.GCSys), true
	case "HeapIdle":
		return float64(s.memStats.HeapIdle), true
	case "HeapInuse":
		return float64(s.memStats.HeapInuse), true
	case "HeapObjects":
		return float64(s.memStats.HeapObjects), true
	case "HeapReleased":
		return float64(s.memStats.HeapReleased), true
	case "LastGC":
		return float64(s.memStats.LastGC), true
	case "Lookups":
		return float64(s.memStats.Lookups), true
	case "MCacheInuse":
		return float64(s.memStats.MCacheInuse), true
	case "MCacheSys":
		return float64(s.memStats.MCacheSys), true
	case "MSpanInuse":
		return float64(s.memStats.MSpanInuse), true
	case "MSpanSys":
		return float64(s.memStats.MSpanSys), true
	case "Mallocs":
		return float64(s.memStats.Mallocs), true
	case "NextGC":
		return float64(s.memStats.NextGC), true
	case "NumForcedGC":
		return float64(s.memStats.NumForcedGC), true
	case "NumGC":
		return float64(s.memStats.NumGC), true
	case "OtherSys":
		return float64(s.memStats.OtherSys), true
	case "PauseTotalNs":
		return float64(s.memStats.PauseTotalNs), true
	case "StackInuse":
		return float64(s.memStats.StackInuse), true
	case "StackSys":
		return float64(s.memStats.StackSys), true
	case "TotalAlloc":
		return float64(s.memStats.TotalAlloc), true
	case "Frees":
		return float64(s.memStats.Frees), true
	case "Sys":
		return float64(s.memStats.Sys), true
	case "RandomValue":
		return float64(s.randomValue), true
	default:
		return 0, false
	}

}

func (s *Stats) GetInt64Value(id string) (int64, bool) {

	switch id {
	case metrics.PollCount:
		return s.pollCount, true
	default:
		return 0, false
	}

}

func (s *Stats) SetFloat64Value(key string, value float64) float64 {
	return 0
}

func (s *Stats) AddInt64Value(key string, value int64) int64 {
	return 0
}

func gaugeMetrics() []string {

	var gauges []string
	gauges = append(gauges, "Alloc")
	gauges = append(gauges, "BuckHashSys")
	gauges = append(gauges, "Frees")
	gauges = append(gauges, "GCCPUFraction")
	gauges = append(gauges, "GCSys")
	gauges = append(gauges, "HeapAlloc")
	gauges = append(gauges, "HeapIdle")
	gauges = append(gauges, "HeapInuse")
	gauges = append(gauges, "HeapObjects")
	gauges = append(gauges, "HeapReleased")
	gauges = append(gauges, "HeapSys")
	gauges = append(gauges, "LastGC")
	gauges = append(gauges, "Lookups")
	gauges = append(gauges, "MCacheInuse")
	gauges = append(gauges, "MCacheSys")
	gauges = append(gauges, "MSpanInuse")
	gauges = append(gauges, "MSpanSys")
	gauges = append(gauges, "Mallocs")
	gauges = append(gauges, "NextGC")
	gauges = append(gauges, "NumForcedGC")
	gauges = append(gauges, "NumGC")
	gauges = append(gauges, "OtherSys")
	gauges = append(gauges, "PauseTotalNs")
	gauges = append(gauges, "StackInuse")
	gauges = append(gauges, "StackSys")
	gauges = append(gauges, "Sys")
	gauges = append(gauges, "TotalAlloc")
	gauges = append(gauges, "RandomValue")

	return gauges

}

func counterMetrics() []string {

	var counters []string
	counters = append(counters, "PollCount")

	return counters

}
