package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v8"

	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

type Config struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Server         string `env:"ADDRESS"`
}

type client interface {
	Push(mType string, Name string, Value string) error
}

func main() {

	var lastReportPush time.Time

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}
	s := stats.NewStats()

	pause := time.Duration(cfg.PollInterval) * time.Second

	for {
		s.Update()
		now := time.Now()
		if isTimeToPushReport(lastReportPush, now, cfg.ReportInterval) {
			conn := metcoll.NewClient(cfg.Server)
			if err := pushReport(conn, s, cfg); err != nil {
				log.Print(err)
			}
			s.ClearPollCount()
		}
		time.Sleep(pause)
	}

}

func pushReport(conn client, s *stats.Stats, cfg *Config) error {

	for mType, data := range s.GetReportData() {

		for name, value := range data {
			if err := conn.Push(mType, name, value); err != nil {
				return fmt.Errorf("cannot push %s %s with value %s on server: %v", mType, name, value, err)
			}
		}

	}

	return nil

}

func parseConfig() (*Config, error) {

	var c Config

	c.Server = *flag.String("a", "localhost:8080", "server end point")
	c.ReportInterval = *flag.Int("r", 10, "report push interval")
	c.PollInterval = *flag.Int("p", 2, "poll interval")
	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	return &c, nil

}

func isTimeToPushReport(lastReportPush time.Time, now time.Time, reportInterval int) bool {
	return now.After(lastReportPush.Add(time.Second * time.Duration(reportInterval)))
}
