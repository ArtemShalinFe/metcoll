package main

import (
	"errors"
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

	for {
		s.Update()
		now := time.Now()
		if isTimeToPushReport(lastReportPush, now, cfg.ReportInterval) {
			conn := metcoll.NewClient(cfg.Server)
			err := pushReport(conn, s, cfg)
			if err != nil {
				log.Print(err)
			}
			s.IncPollCount()
		}
		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
	}

}

func pushReport(conn client, s *stats.Stats, cfg *Config) error {

	for mType, data := range s.GetReportData() {

		for name, value := range data {
			err := conn.Push(mType, name, value)
			if err != nil {
				t := fmt.Sprintf("cannot push %s %s with value %s on server: %v", mType, name, value, err)
				return errors.New(t)
			}
		}

	}

	return nil

}

func parseConfig() (*Config, error) {

	var c Config

	flagServer := flag.String("a", "localhost:8080", "server end point")
	flagReportInterval := flag.Int("r", 10, "report push interval")
	flagPollInterval := flag.Int("p", 2, "poll interval")
	flag.Parse()

	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}

	if c.Server == "" {
		c.Server = *flagServer
	}

	if c.ReportInterval == 0 {
		c.ReportInterval = *flagReportInterval
	}
	if c.PollInterval == 0 {
		c.PollInterval = *flagPollInterval
	}

	return &c, nil

}

func isTimeToPushReport(lastReportPush time.Time, now time.Time, reportInterval int) bool {
	return now.After(lastReportPush.Add(time.Second * time.Duration(reportInterval)))
}
