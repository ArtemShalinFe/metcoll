package configuration

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

const (
	defaultReportinterval  = 10
	reportIntervalFlagName = "r"

	defaultPollInterval  = 2
	pollIntervalFlagName = "p"

	defaultLimit  = 1
	limitFlagName = "l"

	defaultMetcollAddress  = "localhost:8080"
	metcollAddressFlagName = "a"

	defaultHashKey  = ""
	hashKeyFlagName = "k"
)

// ConfigAgent contains configuration for agent.
type ConfigAgent struct {
	Server         string `env:"ADDRESS"`
	Key            []byte
	PollInterval   int `env:"POLL_INTERVAL"`
	ReportInterval int `env:"REPORT_INTERVAL"`
	Limit          int `env:"RATE_LIMIT"`
}

// ParseAgent - return parsed config.
//
// Environment variables have higher priority over command line variables.
func ParseAgent() (*ConfigAgent, error) {
	var c ConfigAgent

	var hashkey string
	flag.StringVar(&c.Server, metcollAddressFlagName, defaultMetcollAddress, "address metcoll server")
	flag.IntVar(&c.ReportInterval, reportIntervalFlagName, defaultReportinterval, "report push interval")
	flag.IntVar(&c.PollInterval, pollIntervalFlagName, defaultPollInterval, "poll interval")
	flag.StringVar(&hashkey, hashKeyFlagName, defaultHashKey, "hash key")
	flag.IntVar(&c.Limit, limitFlagName, defaultLimit, "worker limit")

	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("env parse agent config err: %w", err)
	}

	envkey := os.Getenv(envHashKey)
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
