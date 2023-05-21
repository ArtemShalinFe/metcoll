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

	gaugeMetrics := make(map[string]*metrics.Metrics)

	gaugeMetrics["Alloc"] = metrics.NewGaugeMetric("Alloc", float64(s.memStats.Alloc))
	gaugeMetrics["BuckHashSys"] = metrics.NewGaugeMetric("BuckHashSys", float64(s.memStats.BuckHashSys))
	gaugeMetrics["GCCPUFraction"] = metrics.NewGaugeMetric("GCCPUFraction", float64(s.memStats.GCCPUFraction))
	gaugeMetrics["HeapAlloc"] = metrics.NewGaugeMetric("HeapAlloc", float64(s.memStats.HeapAlloc))
	gaugeMetrics["GCSys"] = metrics.NewGaugeMetric("GCSys", float64(s.memStats.GCSys))
	gaugeMetrics["HeapIdle"] = metrics.NewGaugeMetric("HeapIdle", float64(s.memStats.HeapIdle))
	gaugeMetrics["HeapInuse"] = metrics.NewGaugeMetric("HeapInuse", float64(s.memStats.HeapInuse))
	gaugeMetrics["HeapObjects"] = metrics.NewGaugeMetric("HeapObjects", float64(s.memStats.HeapObjects))
	gaugeMetrics["HeapReleased"] = metrics.NewGaugeMetric("HeapReleased", float64(s.memStats.HeapReleased))
	gaugeMetrics["HeapSys"] = metrics.NewGaugeMetric("HeapSys", float64(s.memStats.HeapSys))
	gaugeMetrics["LastGC"] = metrics.NewGaugeMetric("LastGC", float64(s.memStats.LastGC))
	gaugeMetrics["Lookups"] = metrics.NewGaugeMetric("Lookups", float64(s.memStats.Lookups))
	gaugeMetrics["MCacheInuse"] = metrics.NewGaugeMetric("MCacheInuse", float64(s.memStats.MCacheInuse))
	gaugeMetrics["MCacheSys"] = metrics.NewGaugeMetric("MCacheSys", float64(s.memStats.MCacheSys))
	gaugeMetrics["MSpanInuse"] = metrics.NewGaugeMetric("MSpanInuse", float64(s.memStats.MSpanInuse))
	gaugeMetrics["MSpanSys"] = metrics.NewGaugeMetric("MSpanSys", float64(s.memStats.MSpanSys))
	gaugeMetrics["Mallocs"] = metrics.NewGaugeMetric("Mallocs", float64(s.memStats.Mallocs))
	gaugeMetrics["NextGC"] = metrics.NewGaugeMetric("NextGC", float64(s.memStats.NextGC))
	gaugeMetrics["NumForcedGC"] = metrics.NewGaugeMetric("NumForcedGC", float64(s.memStats.NumForcedGC))
	gaugeMetrics["NumGC"] = metrics.NewGaugeMetric("NumGC", float64(s.memStats.NumGC))
	gaugeMetrics["OtherSys"] = metrics.NewGaugeMetric("OtherSys", float64(s.memStats.OtherSys))
	gaugeMetrics["PauseTotalNs"] = metrics.NewGaugeMetric("PauseTotalNs", float64(s.memStats.PauseTotalNs))
	gaugeMetrics["StackInuse"] = metrics.NewGaugeMetric("StackInuse", float64(s.memStats.StackInuse))
	gaugeMetrics["StackSys"] = metrics.NewGaugeMetric("StackSys", float64(s.memStats.StackSys))
	gaugeMetrics["TotalAlloc"] = metrics.NewGaugeMetric("TotalAlloc", float64(s.memStats.TotalAlloc))
	gaugeMetrics["RandomValue"] = metrics.NewGaugeMetric("RandomValue", float64(s.randomValue))
	gaugeMetrics["Frees"] = metrics.NewGaugeMetric("Frees", float64(s.memStats.Frees))
	gaugeMetrics["Sys"] = metrics.NewGaugeMetric("Sys", float64(s.memStats.Sys))

	counterMetrics := make(map[string]*metrics.Metrics)
	counterMetrics[metrics.PollCount] = metrics.NewCounterMetric(metrics.PollCount, s.pollCount)

	reportData := make(map[string]map[string]*metrics.Metrics)
	reportData[metrics.GaugeMetric] = gaugeMetrics
	reportData[metrics.CounterMetric] = counterMetrics

	return reportData

}
