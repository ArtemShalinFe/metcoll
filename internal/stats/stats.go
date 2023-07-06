package stats

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

type Stats struct {
	mux         *sync.RWMutex
	memStats    *runtime.MemStats
	pollCount   int64
	randomValue int64 // timestamp
}

func NewStats() *Stats {

	return &Stats{
		mux:         &sync.RWMutex{},
		memStats:    &runtime.MemStats{},
		pollCount:   0,
		randomValue: time.Now().Unix(),
	}
}

func (s *Stats) RunCollectBatchStats(ctx context.Context, cfg *configuration.ConfigAgent, ms chan<- []*metrics.Metrics) {

	pauseUpdate := time.Duration(cfg.PollInterval) * time.Second
	pauseCollect := time.Duration(cfg.ReportInterval) * time.Second

	go s.update(pauseUpdate)
	go s.batchCollect(ctx, pauseCollect, ms)

}

func (s *Stats) update(pause time.Duration) {

	if pause == 0 {
		pause = 2 * time.Second
	}

	go func() {

		for {

			s.mux.Lock()

			runtime.ReadMemStats(s.memStats)
			s.randomValue = time.Now().Unix()
			s.pollCount = s.pollCount + 1
			s.mux.Unlock()

			time.Sleep(pause)

		}

	}()

}

func (s *Stats) batchCollect(ctx context.Context, pause time.Duration, ms chan<- []*metrics.Metrics) {

	if pause == 0 {
		pause = 10 * time.Second
	}
	go func() {
		for {

			var mcs []*metrics.Metrics
			for _, data := range s.GetReportData(ctx) {
				for _, metric := range data {
					mcs = append(mcs, metric)
				}
			}

			select {
			case <-ctx.Done():
				return
			case ms <- mcs:
			default:
			}

			time.Sleep(pause)

		}

	}()

}

func (s *Stats) ClearPollCount() {
	s.pollCount = 0
}

func (s *Stats) GetReportData(ctx context.Context) map[string]map[string]*metrics.Metrics {

	gm := make(map[string]*metrics.Metrics)
	for _, id := range gaugeMetrics() {
		m := metrics.NewGaugeMetric(id, 0)
		m.Get(ctx, s)
		gm[id] = m
	}

	cm := make(map[string]*metrics.Metrics)
	for _, mid := range counterMetrics() {
		m := metrics.NewCounterMetric(mid, 0)
		m.Get(ctx, s)
		cm[mid] = m
	}

	reportData := make(map[string]map[string]*metrics.Metrics)
	reportData[metrics.GaugeMetric] = gm
	reportData[metrics.CounterMetric] = cm

	return reportData

}

func (s *Stats) GetFloat64Value(ctx context.Context, id string) (float64, error) {

	s.mux.RLock()
	defer s.mux.RUnlock()

	switch id {
	case "Alloc":
		return float64(s.memStats.Alloc), nil
	case "BuckHashSys":
		return float64(s.memStats.BuckHashSys), nil
	case "GCCPUFraction":
		return float64(s.memStats.GCCPUFraction), nil
	case "HeapAlloc":
		return float64(s.memStats.HeapAlloc), nil
	case "GCSys":
		return float64(s.memStats.GCSys), nil
	case "HeapIdle":
		return float64(s.memStats.HeapIdle), nil
	case "HeapInuse":
		return float64(s.memStats.HeapInuse), nil
	case "HeapObjects":
		return float64(s.memStats.HeapObjects), nil
	case "HeapReleased":
		return float64(s.memStats.HeapReleased), nil
	case "LastGC":
		return float64(s.memStats.LastGC), nil
	case "Lookups":
		return float64(s.memStats.Lookups), nil
	case "MCacheInuse":
		return float64(s.memStats.MCacheInuse), nil
	case "MCacheSys":
		return float64(s.memStats.MCacheSys), nil
	case "MSpanInuse":
		return float64(s.memStats.MSpanInuse), nil
	case "MSpanSys":
		return float64(s.memStats.MSpanSys), nil
	case "Mallocs":
		return float64(s.memStats.Mallocs), nil
	case "NextGC":
		return float64(s.memStats.NextGC), nil
	case "NumForcedGC":
		return float64(s.memStats.NumForcedGC), nil
	case "NumGC":
		return float64(s.memStats.NumGC), nil
	case "OtherSys":
		return float64(s.memStats.OtherSys), nil
	case "PauseTotalNs":
		return float64(s.memStats.PauseTotalNs), nil
	case "StackInuse":
		return float64(s.memStats.StackInuse), nil
	case "StackSys":
		return float64(s.memStats.StackSys), nil
	case "TotalAlloc":
		return float64(s.memStats.TotalAlloc), nil
	case "Frees":
		return float64(s.memStats.Frees), nil
	case "Sys":
		return float64(s.memStats.Sys), nil
	case "RandomValue":
		return float64(s.randomValue), nil
	case "TotalMemory":

		vm, err := mem.VirtualMemory()
		if err != nil {
			return 0, storage.ErrNoRows
		}
		return float64(vm.Total), nil

	case "FreeMemory":

		vm, err := mem.VirtualMemory()
		if err != nil {
			return 0, nil
		}
		return float64(vm.Free), nil

	case "CPUutilization1":

		c, err := cpu.Info()
		if err != nil {
			return 0, nil
		}

		if len(c) > 0 {
			return float64(c[0].Mhz), nil
		} else {
			return 0, nil
		}

	default:
		return 0, nil
	}

}

func (s *Stats) GetInt64Value(ctx context.Context, id string) (int64, error) {

	s.mux.RLock()
	defer s.mux.RUnlock()

	switch id {
	case metrics.PollCount:
		return s.pollCount, nil
	default:
		return 0, storage.ErrNoRows
	}

}

func (s *Stats) SetFloat64Value(ctx context.Context, key string, value float64) (float64, error) {
	return 0, nil
}

func (s *Stats) AddInt64Value(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}

func (s *Stats) BatchAddInt64Value(ctx context.Context, counters map[string]int64) (map[string]int64, []error, error) {
	return nil, nil, nil
}

func (s *Stats) BatchSetFloat64Value(ctx context.Context, counters map[string]float64) (map[string]float64, []error, error) {
	return nil, nil, nil
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
	gauges = append(gauges, "TotalMemory")
	gauges = append(gauges, "FreeMemory")
	gauges = append(gauges, "CPUutilization1")

	return gauges

}

func counterMetrics() []string {

	var counters []string
	counters = append(counters, "PollCount")

	return counters

}
