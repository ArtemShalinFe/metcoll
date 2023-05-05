package main

import (
	"flag"
	"log"
	"time"

	metcoll "github.com/ArtemShalinFe/metcoll/cmd/agent/internal/client"
	metrics "github.com/ArtemShalinFe/metcoll/cmd/agent/internal/stats"
	"github.com/caarlos0/env/v8"
)

var lastReportPush time.Time
var conn *metcoll.Client

type Config struct {
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	Server         string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func main() {

	cfg := parseConfig()
	conn = metcoll.NewClient(cfg.Server)
	var s metrics.Stats

	for {
		s.Update()
		pushReport(s, cfg)
		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
	}

}

func pushReport(s metrics.Stats, cfg *Config) {

	now := time.Now()
	if isTimeToPushReport(now, cfg.ReportInterval) {

		lastReportPush = now

		for mType, data := range s.GetReportData() {

			for name, value := range data {
				conn.Push(mType, name, value)
			}

		}

	}

}

func parseConfig() *Config {

	var c Config

	flagServer := flag.String("a", "", "server end point")
	flagReportInterval := flag.Int("r", 0, "report push interval")
	flagPollInterval := flag.Int("p", 0, "poll interval")
	flag.Parse()

	err := env.Parse(&c)
	if err != nil {
		log.Fatal(err)
	}

	if *flagServer != "" {
		c.Server = *flagServer
	}
	if *flagReportInterval != 0 {
		c.ReportInterval = *flagReportInterval
	}
	if *flagPollInterval != 0 {
		c.PollInterval = *flagPollInterval
	}

	return &c

}

func isTimeToPushReport(now time.Time, reportInterval int) bool {
	return now.After(lastReportPush.Add(time.Second * time.Duration(reportInterval)))
}
