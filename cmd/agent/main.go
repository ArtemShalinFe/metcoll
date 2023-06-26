package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/interrupter"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

type metcollClient interface {
	BatchUpdate(ctx context.Context, ms []*metrics.Metrics) error
}

func main() {

	i := interrupter.NewInterrupters()

	zl, err := zap.NewProduction()
	if err != nil {
		log.Fatal(fmt.Errorf("cannot init zap-logger err: %w ", err))
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot init middleware logger err: %w ", err))
	}

	rl, err := logger.NewRLLogger(sl)
	if err != nil {
		log.Fatal(err)
	}

	i.Use(l.Interrupt)
	i.Run(l.SugaredLogger)

	cfg, err := configuration.ParseAgent()
	if err != nil {
		l.Errorf("cannot parse server config file err: %w", err)
		return
	}

	l.Infof("parsed agent config: %+v", cfg)

	ctx := context.Background()

	var lastReportPush time.Time
	s := stats.NewStats()
	pause := time.Duration(cfg.PollInterval) * time.Second
	durReportInterval := time.Duration(cfg.ReportInterval) * time.Second
	conn := metcoll.NewClient(cfg.Server, rl)
	for {

		s.Update()
		now := time.Now()

		if isTimeToPushReport(lastReportPush, now, durReportInterval) {

			var ms []*metrics.Metrics
			for _, data := range s.GetReportData(ctx) {
				for _, metric := range data {
					ms = append(ms, metric)
				}
			}

			if len(ms) > 0 {
				if err := pushReport(ctx, conn, ms); err != nil {
					l.Infof("cannot push report on server err: %w", err)
				} else {
					lastReportPush = now
				}
				s.ClearPollCount()
			}

		}

		time.Sleep(pause)

	}

}

func pushReport(ctx context.Context, conn metcollClient, ms []*metrics.Metrics) error {

	if err := conn.BatchUpdate(ctx, ms); err != nil {
		return fmt.Errorf("cannot push batch on server err: %w", err)
	}

	return nil

}

func isTimeToPushReport(lastReportPush time.Time, now time.Time, d time.Duration) bool {
	return now.After(lastReportPush.Add(d))
}
