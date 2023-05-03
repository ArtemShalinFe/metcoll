package main

import (
	"fmt"
	"runtime"
	"time"
)

const gauge = "gauge"
const counter = "counter"

type Metric struct {
	name  string
	value string
	mType string
}

func NewMetric(Name string, Value string, mType string) *Metric {
	return &Metric{
		name:  Name,
		value: Value,
		mType: mType,
	}
}

func GetMetricList(memStats *runtime.MemStats) []Metric {

	var metrics []Metric
	metrics = append(metrics, *NewMetric("Alloc", fmt.Sprintf("%v", memStats.Alloc), gauge))
	metrics = append(metrics, *NewMetric("BuckHashSys", fmt.Sprintf("%v", memStats.BuckHashSys), gauge))
	metrics = append(metrics, *NewMetric("GCCPUFraction", fmt.Sprintf("%v", memStats.GCCPUFraction), gauge))
	metrics = append(metrics, *NewMetric("GCSys", fmt.Sprintf("%v", memStats.GCSys), gauge))
	metrics = append(metrics, *NewMetric("HeapAlloc", fmt.Sprintf("%v", memStats.HeapAlloc), gauge))
	metrics = append(metrics, *NewMetric("HeapIdle", fmt.Sprintf("%v", memStats.HeapIdle), gauge))
	metrics = append(metrics, *NewMetric("HeapInuse", fmt.Sprintf("%v", memStats.HeapInuse), gauge))
	metrics = append(metrics, *NewMetric("HeapObjects", fmt.Sprintf("%v", memStats.HeapObjects), gauge))
	metrics = append(metrics, *NewMetric("HeapReleased", fmt.Sprintf("%v", memStats.HeapReleased), gauge))
	metrics = append(metrics, *NewMetric("HeapSys", fmt.Sprintf("%v", memStats.HeapSys), gauge))
	metrics = append(metrics, *NewMetric("LastGC", fmt.Sprintf("%v", memStats.LastGC), gauge))
	metrics = append(metrics, *NewMetric("Lookups", fmt.Sprintf("%v", memStats.Lookups), gauge))
	metrics = append(metrics, *NewMetric("MCacheInuse", fmt.Sprintf("%v", memStats.MCacheInuse), gauge))
	metrics = append(metrics, *NewMetric("MCacheSys", fmt.Sprintf("%v", memStats.MCacheSys), gauge))
	metrics = append(metrics, *NewMetric("MSpanInuse", fmt.Sprintf("%v", memStats.MSpanInuse), gauge))
	metrics = append(metrics, *NewMetric("MSpanSys", fmt.Sprintf("%v", memStats.MSpanSys), gauge))
	metrics = append(metrics, *NewMetric("Mallocs", fmt.Sprintf("%v", memStats.Mallocs), gauge))
	metrics = append(metrics, *NewMetric("NextGC", fmt.Sprintf("%v", memStats.NextGC), gauge))
	metrics = append(metrics, *NewMetric("NumForcedGC", fmt.Sprintf("%v", memStats.NumForcedGC), gauge))
	metrics = append(metrics, *NewMetric("NumGC", fmt.Sprintf("%v", memStats.NumGC), gauge))
	metrics = append(metrics, *NewMetric("OtherSys", fmt.Sprintf("%v", memStats.OtherSys), gauge))
	metrics = append(metrics, *NewMetric("PauseTotalNs", fmt.Sprintf("%v", memStats.PauseTotalNs), gauge))
	metrics = append(metrics, *NewMetric("StackInuse", fmt.Sprintf("%v", memStats.StackInuse), gauge))
	metrics = append(metrics, *NewMetric("StackSys", fmt.Sprintf("%v", memStats.StackSys), gauge))
	metrics = append(metrics, *NewMetric("Sys", fmt.Sprintf("%v", memStats.Sys), gauge))
	metrics = append(metrics, *NewMetric("TotalAlloc", fmt.Sprintf("%v", memStats.TotalAlloc), gauge))
	metrics = append(metrics, *NewMetric("PollCount", fmt.Sprintf("%v", ms.PollCount), counter))
	metrics = append(metrics, *NewMetric("RandomValue", fmt.Sprintf("%v", time.Now().Unix()), gauge))

	return metrics

}
