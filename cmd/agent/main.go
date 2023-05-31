package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v8"

	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

type metcollClient interface {
	Update(m *metrics.Metrics) error
}

type Config struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Server         string `env:"ADDRESS"`
}

func main() {

	var lastReportPush time.Time
	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()

	cfg, err := parseConfig()
	if err != nil {
		l.Error("cannot parse agent config file err: ", err)
		return
	}

	l.Info("parsed agent config: ", fmt.Sprintf("%+v", cfg))

	s := stats.NewStats()

	pause := time.Duration(cfg.PollInterval) * time.Second
	durReportInterval := time.Duration(cfg.ReportInterval) * time.Second
	conn := metcoll.NewClient(cfg.Server, l)

	for {

		s.Update()
		now := time.Now()

		if isTimeToPushReport(lastReportPush, now, durReportInterval) {

			if err := pushReport(conn, s, cfg); err != nil {
				l.Info(err)
			} else {
				lastReportPush = now
			}

		}

		time.Sleep(pause)

	}

}

func pushReport(conn metcollClient, s *stats.Stats, cfg *Config) error {

	for mType, data := range s.GetReportData() {

		for name, metric := range data {

			if err := conn.Update(metric); err != nil {
				return fmt.Errorf("cannot push %s %s with value %v on server err: %w", mType, name, metric, err)
			}

			if metric.IsPollCount() {
				s.ClearPollCount()
			}

		}

	}

	return nil

}

func parseConfig() (*Config, error) {

	var c Config

	flag.StringVar(&c.Server, "a", "localhost:8080", "server end point")
	flag.IntVar(&c.ReportInterval, "r", 10, "report push interval")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval")
	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	return &c, nil

}

func isTimeToPushReport(lastReportPush time.Time, now time.Time, d time.Duration) bool {
	return now.After(lastReportPush.Add(d))
}
