package main

import (
	"flag"
	"runtime"
	"time"

	metcoll "github.com/ArtemShalinFe/metcoll/internal/client"
)

var pollInterval int
var reportInterval int

var lastReportPush time.Time
var ms *lastMemStat
var conn *metcoll.Client

type lastMemStat struct {
	lms         *runtime.MemStats
	PollCount   int64
	RandomValue int64 // timestamp
}

func main() {

	serverEndPoint := flag.String("a", "localhost:8080", "server end point")
	ri := flag.Int("r", 10, "report push interval")
	pi := flag.Int("p", 2, "poll interval")

	flag.Parse()

	conn = metcoll.NewClient(*serverEndPoint)

	reportInterval = *ri
	pollInterval = *pi

	for {

		runtime.ReadMemStats(ms.lms)
		ms.PollCount += 1

		now := time.Now()
		if isTimeToPushReport(now) {
			lastReportPush = now

			for _, metric := range GetMetricList(ms.lms) {
				conn.Push(metric.mType, metric.name, metric.value)
			}

		}

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}

}

func init() {

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	ms = &lastMemStat{
		lms:         &m,
		PollCount:   0,
		RandomValue: time.Now().Unix(),
	}

}

func isTimeToPushReport(now time.Time) bool {
	return now.After(lastReportPush.Add(time.Second * time.Duration(reportInterval)))
}
