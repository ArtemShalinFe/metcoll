package configuration

import (
	"flag"

	"github.com/caarlos0/env"
)

type ConfigAgent struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Server         string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	Limit          int    `env:"RATE_LIMIT"`
	HashKey        []byte
}

func ParseAgent() (*ConfigAgent, error) {

	var c ConfigAgent

	flag.StringVar(&c.Server, "a", "localhost:8080", "server end point")
	flag.IntVar(&c.ReportInterval, "r", 10, "report push interval")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&c.Key, "k", "", "hash key")
	flag.IntVar(&c.Limit, "l", 1, "limit")

	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	c.HashKey = []byte(c.Key)

	return &c, nil

}
