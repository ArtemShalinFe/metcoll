package main

import (
	"context"
	"log"
	"time"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/interrupter"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

func main() {

	zl, err := zap.NewProduction()
	if err != nil {
		log.Fatal(fmt.Errorf("cannot init zap-logger err: %w ", err))
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		log.Fatal(err)
	}

	i := interrupter.NewInterrupters()
	i.Use(l.Interrupt)
	i.Run(l.SugaredLogger)

	cfg, err := configuration.ParseAgent()
	if err != nil {
		l.Errorf("cannot parse server config file err: %w", err)
		return
	}

	l.Infof("parsed agent config: %+v", cfg)

	rl, err := logger.NewRLLogger(sl)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client := metcoll.NewClient(cfg, l)
	stats := stats.NewStats()

	if cfg.Limit > 0 {
	if cfg.Limit > 0 {

		s.Update()
		now := time.Now()

		if isTimeToPushReport(lastReportPush, now, durReportInterval) {

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

		time.Sleep(pause)

		for err := range errs {
			if err != nil {
				l.Log.Errorf("batch update metrics failed err: %w", err)
			} else {
				stats.ClearPollCount()
			}
		}
	}

}
