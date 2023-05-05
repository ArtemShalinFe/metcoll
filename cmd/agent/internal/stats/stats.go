package stats

import (
	"fmt"
	"runtime"
	"time"
)

const gauge = "gauge"
const counter = "counter"

type Stats struct {
	memStats    *runtime.MemStats
	PollCount   int64
	RandomValue int64 // timestamp
}

func (s *Stats) Update() {

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	s.memStats = &m
	s.PollCount = s.PollCount + 1
	s.RandomValue = time.Now().Unix()

}

func (s *Stats) GetReportData() map[string]map[string]string {

	gaugeMetrics := make(map[string]string)
	counterMetrics := make(map[string]string)

	reportData := make(map[string]map[string]string)

	gaugeMetrics["Alloc"] = fmt.Sprintf("%v", s.memStats.Alloc)
	gaugeMetrics["BuckHashSys"] = fmt.Sprintf("%v", s.memStats.BuckHashSys)
	gaugeMetrics["GCCPUFraction"] = fmt.Sprintf("%v", s.memStats.GCCPUFraction)
	gaugeMetrics["HeapAlloc"] = fmt.Sprintf("%v", s.memStats.HeapAlloc)
	gaugeMetrics["GCSys"] = fmt.Sprintf("%v", s.memStats.GCSys)
	gaugeMetrics["HeapIdle"] = fmt.Sprintf("%v", s.memStats.HeapIdle)
	gaugeMetrics["HeapInuse"] = fmt.Sprintf("%v", s.memStats.HeapInuse)
	gaugeMetrics["HeapObjects"] = fmt.Sprintf("%v", s.memStats.HeapObjects)
	gaugeMetrics["HeapReleased"] = fmt.Sprintf("%v", s.memStats.HeapReleased)
	gaugeMetrics["HeapSys"] = fmt.Sprintf("%v", s.memStats.HeapSys)
	gaugeMetrics["LastGC"] = fmt.Sprintf("%v", s.memStats.LastGC)
	gaugeMetrics["Lookups"] = fmt.Sprintf("%v", s.memStats.Lookups)
	gaugeMetrics["MCacheInuse"] = fmt.Sprintf("%v", s.memStats.MCacheInuse)
	gaugeMetrics["MCacheSys"] = fmt.Sprintf("%v", s.memStats.MCacheSys)
	gaugeMetrics["MSpanInuse"] = fmt.Sprintf("%v", s.memStats.MSpanInuse)
	gaugeMetrics["MSpanSys"] = fmt.Sprintf("%v", s.memStats.MSpanSys)
	gaugeMetrics["Mallocs"] = fmt.Sprintf("%v", s.memStats.Mallocs)
	gaugeMetrics["NextGC"] = fmt.Sprintf("%v", s.memStats.NextGC)
	gaugeMetrics["NumForcedGC"] = fmt.Sprintf("%v", s.memStats.NumForcedGC)
	gaugeMetrics["NumGC"] = fmt.Sprintf("%v", s.memStats.NumGC)
	gaugeMetrics["OtherSys"] = fmt.Sprintf("%v", s.memStats.OtherSys)
	gaugeMetrics["PauseTotalNs"] = fmt.Sprintf("%v", s.memStats.PauseTotalNs)
	gaugeMetrics["StackInuse"] = fmt.Sprintf("%v", s.memStats.StackInuse)
	gaugeMetrics["StackSys"] = fmt.Sprintf("%v", s.memStats.StackSys)
	gaugeMetrics["TotalAlloc"] = fmt.Sprintf("%v", s.memStats.TotalAlloc)
	gaugeMetrics["RandomValue"] = fmt.Sprintf("%v", s.RandomValue)

	counterMetrics["PollCount"] = fmt.Sprintf("%v", s.PollCount)

	reportData[gauge] = gaugeMetrics
	reportData[counter] = counterMetrics

	return reportData

}
