package main

import (
	"fmt"
	"runtime"
	"time"
)

const pollInterval = 2
const reportInterval = 10

var lastReportPush time.Time
var ms *lastMemStat
var conn *Server

type lastMemStat struct {
	lms         *runtime.MemStats
	PollCount   int64
	RandomValue int64 // timestamp
}

func main() {

	for {

		runtime.ReadMemStats(ms.lms)
		ms.PollCount += 1

		now := time.Now()
		if isTimeToPushReport(now) {
			lastReportPush = now
			pushvalues()
		}

		time.Sleep(pollInterval * time.Second)
	}

}

func init() {

	conn = &Server{
		host: "localhost",
		port: "8080",
	}

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	ms = &lastMemStat{
		lms:         &m,
		PollCount:   0,
		RandomValue: time.Now().Unix(),
	}

}

func isTimeToPushReport(now time.Time) bool {
	return now.After(lastReportPush.Add(time.Second * reportInterval))
}

// Извиняюсь. Это жесть какая-то, но я не придумал как сделать проще :(
func pushvalues() {

	var metrics []*Metric
	metrics = append(metrics, NewMetric("Alloc", fmt.Sprintf("%v", ms.lms.Alloc), gauge))
	metrics = append(metrics, NewMetric("BuckHashSys", fmt.Sprintf("%v", ms.lms.BuckHashSys), gauge))
	metrics = append(metrics, NewMetric("GCCPUFraction", fmt.Sprintf("%v", ms.lms.GCCPUFraction), gauge))
	metrics = append(metrics, NewMetric("GCSys", fmt.Sprintf("%v", ms.lms.GCSys), gauge))
	metrics = append(metrics, NewMetric("HeapAlloc", fmt.Sprintf("%v", ms.lms.HeapAlloc), gauge))
	metrics = append(metrics, NewMetric("HeapIdle", fmt.Sprintf("%v", ms.lms.HeapIdle), gauge))
	metrics = append(metrics, NewMetric("HeapInuse", fmt.Sprintf("%v", ms.lms.HeapInuse), gauge))
	metrics = append(metrics, NewMetric("HeapObjects", fmt.Sprintf("%v", ms.lms.HeapObjects), gauge))
	metrics = append(metrics, NewMetric("HeapReleased", fmt.Sprintf("%v", ms.lms.HeapReleased), gauge))
	metrics = append(metrics, NewMetric("HeapSys", fmt.Sprintf("%v", ms.lms.HeapSys), gauge))
	metrics = append(metrics, NewMetric("LastGC", fmt.Sprintf("%v", ms.lms.LastGC), gauge))
	metrics = append(metrics, NewMetric("Lookups", fmt.Sprintf("%v", ms.lms.Lookups), gauge))
	metrics = append(metrics, NewMetric("MCacheInuse", fmt.Sprintf("%v", ms.lms.MCacheInuse), gauge))
	metrics = append(metrics, NewMetric("MCacheSys", fmt.Sprintf("%v", ms.lms.MCacheSys), gauge))
	metrics = append(metrics, NewMetric("MSpanInuse", fmt.Sprintf("%v", ms.lms.MSpanInuse), gauge))
	metrics = append(metrics, NewMetric("MSpanSys", fmt.Sprintf("%v", ms.lms.MSpanSys), gauge))
	metrics = append(metrics, NewMetric("Mallocs", fmt.Sprintf("%v", ms.lms.Mallocs), gauge))
	metrics = append(metrics, NewMetric("NextGC", fmt.Sprintf("%v", ms.lms.NextGC), gauge))
	metrics = append(metrics, NewMetric("NumForcedGC", fmt.Sprintf("%v", ms.lms.NumForcedGC), gauge))
	metrics = append(metrics, NewMetric("NumGC", fmt.Sprintf("%v", ms.lms.NumGC), gauge))
	metrics = append(metrics, NewMetric("OtherSys", fmt.Sprintf("%v", ms.lms.OtherSys), gauge))
	metrics = append(metrics, NewMetric("PauseTotalNs", fmt.Sprintf("%v", ms.lms.PauseTotalNs), gauge))
	metrics = append(metrics, NewMetric("StackInuse", fmt.Sprintf("%v", ms.lms.StackInuse), gauge))
	metrics = append(metrics, NewMetric("StackSys", fmt.Sprintf("%v", ms.lms.StackSys), gauge))
	metrics = append(metrics, NewMetric("Sys", fmt.Sprintf("%v", ms.lms.Sys), gauge))
	metrics = append(metrics, NewMetric("TotalAlloc", fmt.Sprintf("%v", ms.lms.TotalAlloc), gauge))
	metrics = append(metrics, NewMetric("PollCount", fmt.Sprintf("%v", ms.PollCount), counter))
	metrics = append(metrics, NewMetric("RandomValue", fmt.Sprintf("%v", time.Now().Unix()), gauge))

	for _, metric := range metrics {
		metric.Push(conn)
	}

}
