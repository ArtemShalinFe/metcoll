// Package stats contains methods and functions for collecting metrics.
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

const (
	Alloc           = "Alloc"
	BuckHashSys     = "BuckHashSys"
	Frees           = "Frees"
	GCCPUFraction   = "GCCPUFraction"
	GCSys           = "GCSys"
	HeapAlloc       = "HeapAlloc"
	HeapIdle        = "HeapIdle"
	HeapInuse       = "HeapInuse"
	HeapObjects     = "HeapObjects"
	HeapReleased    = "HeapReleased"
	HeapSys         = "HeapSys"
	LastGC          = "LastGC"
	Lookups         = "Lookups"
	MCacheInuse     = "MCacheInuse"
	MCacheSys       = "MCacheSys"
	MSpanInuse      = "MSpanInuse"
	MSpanSys        = "MSpanSys"
	Mallocs         = "Mallocs"
	NextGC          = "NextGC"
	NumForcedGC     = "NumForcedGC"
	NumGC           = "NumGC"
	OtherSys        = "OtherSys"
	PauseTotalNs    = "PauseTotalNs"
	StackInuse      = "StackInuse"
	StackSys        = "StackSys"
	Sys             = "Sys"
	TotalAlloc      = "TotalAlloc"
	RandomValue     = "RandomValue"
	TotalMemory     = "TotalMemory"
	FreeMemory      = "FreeMemory"
	CPUutilization1 = "CPUutilization1"
	PollCount       = "PollCount"
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

func (s *Stats) RunCollectBatchStats(ctx context.Context,
	cfg *configuration.ConfigAgent, ms chan<- []*metrics.Metrics) {
	pauseUpdate := time.Duration(cfg.PollInterval) * time.Second
	pauseCollect := time.Duration(cfg.ReportInterval) * time.Second

	go s.update(pauseUpdate)
	go s.batchCollect(ctx, pauseCollect, ms)
}

func (s *Stats) update(pause time.Duration) {
	if pause == 0 {
		const defaultPause = 2 * time.Second
		pause = defaultPause
	}

	for {
		s.mux.Lock()

		runtime.ReadMemStats(s.memStats)
		s.randomValue = time.Now().Unix()
		s.pollCount++

		s.mux.Unlock()

		time.Sleep(pause)
	}
}

func (s *Stats) batchCollect(ctx context.Context, pause time.Duration, ms chan<- []*metrics.Metrics) {
	if pause == 0 {
		const defaultPause = 10 * time.Second
		pause = defaultPause
	}

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
	case Alloc:
		return float64(s.memStats.Alloc), nil
	case BuckHashSys:
		return float64(s.memStats.BuckHashSys), nil
	case GCCPUFraction:
		return float64(s.memStats.GCCPUFraction), nil
	case HeapAlloc:
		return float64(s.memStats.HeapAlloc), nil
	case GCSys:
		return float64(s.memStats.GCSys), nil
	case HeapIdle:
		return float64(s.memStats.HeapIdle), nil
	case HeapInuse:
		return float64(s.memStats.HeapInuse), nil
	case HeapObjects:
		return float64(s.memStats.HeapObjects), nil
	case HeapReleased:
		return float64(s.memStats.HeapReleased), nil
	case LastGC:
		return float64(s.memStats.LastGC), nil
	case Lookups:
		return float64(s.memStats.Lookups), nil
	case MCacheInuse:
		return float64(s.memStats.MCacheInuse), nil
	case MCacheSys:
		return float64(s.memStats.MCacheSys), nil
	case MSpanInuse:
		return float64(s.memStats.MSpanInuse), nil
	case MSpanSys:
		return float64(s.memStats.MSpanSys), nil
	case Mallocs:
		return float64(s.memStats.Mallocs), nil
	case NextGC:
		return float64(s.memStats.NextGC), nil
	case NumForcedGC:
		return float64(s.memStats.NumForcedGC), nil
	case NumGC:
		return float64(s.memStats.NumGC), nil
	case OtherSys:
		return float64(s.memStats.OtherSys), nil
	case PauseTotalNs:
		return float64(s.memStats.PauseTotalNs), nil
	case StackInuse:
		return float64(s.memStats.StackInuse), nil
	case StackSys:
		return float64(s.memStats.StackSys), nil
	case TotalAlloc:
		return float64(s.memStats.TotalAlloc), nil
	case Frees:
		return float64(s.memStats.Frees), nil
	case Sys:
		return float64(s.memStats.Sys), nil
	case RandomValue:
		return float64(s.randomValue), nil
	case TotalMemory:
		vm, err := mem.VirtualMemory()
		if err != nil {
			return 0, storage.ErrNoRows
		}
		return float64(vm.Total), nil
	case FreeMemory:
		vm, err := mem.VirtualMemory()
		if err != nil {
			return 0, err
		}
		return float64(vm.Free), nil
	case CPUutilization1:
		c, err := cpu.Info()
		if err != nil {
			return 0, err
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

func (s *Stats) BatchAddInt64Value(ctx context.Context,
	counters map[string]int64) (map[string]int64, []error, error) {
	return nil, nil, nil
}

func (s *Stats) BatchSetFloat64Value(ctx context.Context,
	counters map[string]float64) (map[string]float64, []error, error) {
	return nil, nil, nil
}

func gaugeMetrics() []string {
	var gauges []string
	gauges = append(gauges,
		Alloc,
		BuckHashSys,
		Frees,
		GCCPUFraction,
		GCSys,
		HeapAlloc,
		HeapIdle,
		HeapInuse,
		HeapObjects,
		HeapReleased,
		HeapSys,
		LastGC,
		Lookups,
		MCacheInuse,
		MCacheSys,
		MSpanInuse,
		MSpanSys,
		Mallocs,
		NextGC,
		NumForcedGC,
		NumGC,
		OtherSys,
		PauseTotalNs,
		StackInuse,
		StackSys,
		Sys,
		TotalAlloc,
		RandomValue,
		TotalMemory,
		FreeMemory,
		CPUutilization1,
	)

	return gauges
}

func counterMetrics() []string {
	var counters []string
	counters = append(counters, PollCount)

	return counters
}
