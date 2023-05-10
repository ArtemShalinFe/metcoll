package stats

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

const gaugeMetric = "gauge"
const counterMetric = "counter"

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

}

func (s *Stats) IncPollCount() {
	s.pollCount = s.pollCount + 1
}

func (s *Stats) GetReportData() map[string]map[string]string {

	gaugeMetrics := make(map[string]string)
	gaugeMetrics["Alloc"] = strconv.FormatInt(int64(s.memStats.Alloc), 10)
	gaugeMetrics["BuckHashSys"] = strconv.FormatInt(int64(s.memStats.BuckHashSys), 10)
	gaugeMetrics["GCCPUFraction"] = strconv.FormatInt(int64(s.memStats.GCCPUFraction), 10)
	gaugeMetrics["HeapAlloc"] = strconv.FormatInt(int64(s.memStats.HeapAlloc), 10)
	gaugeMetrics["GCSys"] = strconv.FormatInt(int64(s.memStats.GCSys), 10)
	gaugeMetrics["HeapIdle"] = strconv.FormatInt(int64(s.memStats.HeapIdle), 10)
	gaugeMetrics["HeapInuse"] = strconv.FormatInt(int64(s.memStats.HeapInuse), 10)
	gaugeMetrics["HeapObjects"] = strconv.FormatInt(int64(s.memStats.HeapObjects), 10)
	gaugeMetrics["HeapReleased"] = strconv.FormatInt(int64(s.memStats.HeapReleased), 10)
	gaugeMetrics["HeapSys"] = strconv.FormatInt(int64(s.memStats.HeapSys), 10)
	gaugeMetrics["LastGC"] = strconv.FormatInt(int64(s.memStats.LastGC), 10)
	gaugeMetrics["Lookups"] = strconv.FormatInt(int64(s.memStats.Lookups), 10)
	gaugeMetrics["MCacheInuse"] = strconv.FormatInt(int64(s.memStats.MCacheInuse), 10)
	gaugeMetrics["MCacheSys"] = strconv.FormatInt(int64(s.memStats.MCacheSys), 10)
	gaugeMetrics["MSpanInuse"] = strconv.FormatInt(int64(s.memStats.MSpanInuse), 10)
	gaugeMetrics["MSpanSys"] = strconv.FormatInt(int64(s.memStats.MSpanSys), 10)
	gaugeMetrics["Mallocs"] = strconv.FormatInt(int64(s.memStats.Mallocs), 10)
	gaugeMetrics["NextGC"] = strconv.FormatInt(int64(s.memStats.NextGC), 10)
	gaugeMetrics["NumForcedGC"] = strconv.FormatInt(int64(s.memStats.NumForcedGC), 10)
	gaugeMetrics["NumGC"] = strconv.FormatInt(int64(s.memStats.NumGC), 10)
	gaugeMetrics["OtherSys"] = strconv.FormatInt(int64(s.memStats.OtherSys), 10)
	gaugeMetrics["PauseTotalNs"] = strconv.FormatInt(int64(s.memStats.PauseTotalNs), 10)
	gaugeMetrics["StackInuse"] = strconv.FormatInt(int64(s.memStats.StackInuse), 10)
	gaugeMetrics["StackSys"] = strconv.FormatInt(int64(s.memStats.StackSys), 10)
	gaugeMetrics["TotalAlloc"] = strconv.FormatInt(int64(s.memStats.TotalAlloc), 10)
	gaugeMetrics["RandomValue"] = strconv.FormatInt(s.randomValue, 10)

	counterMetrics := make(map[string]string)
	counterMetrics["PollCount"] = fmt.Sprintf("%v", s.pollCount)

	reportData := make(map[string]map[string]string)
	reportData[gaugeMetric] = gaugeMetrics
	reportData[counterMetric] = counterMetrics

	return reportData

}
