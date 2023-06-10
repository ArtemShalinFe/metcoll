package main

import (
	"context"
	"fmt"
	"log"
	"time"

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

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	i.Use(l.Interrupt)
	i.Run(l)

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
	conn := metcoll.NewClient(cfg.Server, l)
	for {

		s.Update()
		now := time.Now()

		if isTimeToPushReport(lastReportPush, now, durReportInterval) {

			if err := pushReport(ctx, conn, s, cfg); err != nil {
				l.Info(err)
			} else {
				lastReportPush = now
			}

		}

		time.Sleep(pause)

	}

}

func pushReport(ctx context.Context, conn metcollClient, s *stats.Stats, cfg *configuration.ConfigAgent) error {

	var ms []*metrics.Metrics

	for _, data := range s.GetReportData(ctx) {
		for _, metric := range data {
			ms = append(ms, metric)
		}
	}

	if len(ms) > 0 {
		if err := conn.BatchUpdate(ctx, ms); err != nil {
			return fmt.Errorf("cannot push batch on server err: %w", err)
		}
		s.ClearPollCount()
	}

	return nil

}

func isTimeToPushReport(lastReportPush time.Time, now time.Time, d time.Duration) bool {
	return now.After(lastReportPush.Add(d))
}
