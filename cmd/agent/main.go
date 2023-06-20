package main

import (
	"context"
	"log"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/interrupter"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

func main() {

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	i := interrupter.NewInterrupters()
	i.Use(l.Interrupt)
	i.Run(l)

	cfg, err := configuration.ParseAgent()
	if err != nil {
		l.Log.Errorf("cannot parse server config file err: %w", err)
		return
	}

	l.Log.Infof("parsed agent config: %+v", cfg)

	ctx := context.Background()
	client := metcoll.NewClient(cfg, l)
	stats := stats.NewStats()

	if cfg.Limit > 0 {

		ms := make(chan *metrics.Metrics, cfg.Limit)
		defer close(ms)

		prs := make(chan metcoll.PushResult, cfg.Limit)
		defer close(prs)
		stats.RunCollectStats(ctx, cfg, ms)

		for i := 0; i < cfg.Limit; i++ {
			go client.UpdateMetric(ctx, ms, prs)
		}

		for pr := range prs {

			if pr.Err != nil {
				l.Log.Errorf("update metric failed err: %w", pr.Err)
			}

			if pr.Metric.IsPollCount() {
				stats.ClearPollCount()
			}

		}

	} else {

		mcs := make(chan []*metrics.Metrics, 1)
		defer close(mcs)

		errs := make(chan error, 1)
		defer close(errs)

		stats.RunCollectBatchStats(ctx, cfg, mcs)

		go client.BatchUpdateMetric(ctx, mcs, errs)

		for err := range errs {
			if err != nil {
				l.Log.Errorf("batch update metrics failed err: %w", err)
			} else {
				stats.ClearPollCount()
			}
		}
	}

}
