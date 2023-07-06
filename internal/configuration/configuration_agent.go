package configuration

import (
	"flag"
	"os"

	"github.com/caarlos0/env"
)

type ConfigAgent struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Server         string `env:"ADDRESS"`
	Limit          int    `env:"RATE_LIMIT"`
	Key            []byte
}

func ParseAgent() (*ConfigAgent, error) {

	var c ConfigAgent

	var hashkey string
	flag.StringVar(&c.Server, "a", "localhost:8080", "server end point")
	flag.IntVar(&c.ReportInterval, "r", 10, "report push interval")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&hashkey, "k", "", "hash key")
	flag.IntVar(&c.Limit, "l", 1, "limit")

	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	if hashkey == "" {
		hashkey = os.Getenv("KEY")
	}
	c.Key = []byte(hashkey)

	return &c, nil

}
