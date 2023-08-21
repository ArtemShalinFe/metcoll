package configuration

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

// ConfigAgent contains configuration for agent.
type ConfigAgent struct {
	Server         string `env:"ADDRESS"`
	Key            []byte
	PollInterval   int `env:"POLL_INTERVAL"`
	ReportInterval int `env:"REPORT_INTERVAL"`
	Limit          int `env:"RATE_LIMIT"`
}

const defaultReportinterval = 10
const defaultPollInterval = 2
const defaultLimit = 1
const defaultMetcollAddress = "localhost:8080"

// ParseAgent - return parsed config.
//
// Environment variables have higher priority over command line variables.
func ParseAgent() (*ConfigAgent, error) {
	var c ConfigAgent

	var hashkey string
	flag.StringVar(&c.Server, "a", defaultMetcollAddress, "address metcoll server")
	flag.IntVar(&c.ReportInterval, "r", defaultReportinterval, "report push interval")
	flag.IntVar(&c.PollInterval, "p", defaultPollInterval, "poll interval")
	flag.StringVar(&hashkey, "k", "", "hash key")
	flag.IntVar(&c.Limit, "l", defaultLimit, "worker limit")

	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("env parse agent config err: %w", err)
	}

	envkey := os.Getenv("KEY")
	if envkey != "" {
		hashkey = envkey
	}

	c.Key = []byte(hashkey)

	return &c, nil
}

func (c *ConfigAgent) String() string {
	return fmt.Sprintf("Addres: %s, ReportInterval: %d, PollInterval: %d, Limit: %d",
		c.Server, c.ReportInterval, c.PollInterval, c.Limit)
}
