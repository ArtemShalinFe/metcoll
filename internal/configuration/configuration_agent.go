package configuration

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

const (
	defaultReportInterval  = 10
	reportIntervalFlagName = "r"

	defaultPollInterval  = 2
	pollIntervalFlagName = "p"

	defaultLimit  = 1
	limitFlagName = "l"

	defaultMetcollAddress  = "localhost:8080"
	metcollAddressFlagName = "a"

	defaultHashKey  = ""
	hashKeyFlagName = "k"

	defaultConfigPath = ""
	configFlagName    = "c"

	cryptoKeyFlagName    = "crypto-key"
	defaultCryptoKeyPath = ""
)

// ConfigAgent contains configuration for agent.
type ConfigAgent struct {
	Server          string `env:"ADDRESS" json:"address,omitempty"`
	Path            string `env:"CONFIG"`
	PublicCryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	Key             []byte
	PollInterval    int `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
	ReportInterval  int `env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
	Limit           int `env:"RATE_LIMIT"`
}

func newConfigAgent() *ConfigAgent {
	return &ConfigAgent{
		Server:         defaultMetcollAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		Limit:          defaultLimit,
		Path:           defaultConfigPath,
	}
}

// ParseAgent - return parsed config.
//
// Environment variables have higher priority over command line variables and config file.
// Command line variables have higher priority over variables from config file.
func ParseAgent() (*ConfigAgent, error) {
	configCL := readConfigAgentFromCL()

	configENV, err := readConfigAgentFromENV()
	if err != nil {
		return nil, fmt.Errorf("an error occurred when reading the agent configuration env var, err: %w", err)
	}

	path := getConfigVar(configCL.Path, configENV.Path, "", "", "")
	configFile, err := readConfigAgentFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when reading the agent configuration file, err: %w", err)
	}

	var c ConfigAgent
	c.setFromConfigs(configCL, configENV, configFile, path)

	return &c, nil
}

// setFromConfigs -  sets configuration values from instances obtained
// from command line variables, environment variables, configuration file variables.
func (c *ConfigAgent) setFromConfigs(configCL, configENV, configFile *ConfigAgent, path string) {
	c.Server = getConfigVar(
		configCL.Server, configENV.Server, configFile.Server, defaultMetcollAddress, "")

	c.PollInterval = getConfigVar(
		configCL.PollInterval, configENV.PollInterval, configFile.PollInterval, defaultPollInterval, 0)

	c.ReportInterval = getConfigVar(
		configCL.ReportInterval, configENV.ReportInterval, configFile.ReportInterval, defaultReportInterval, 0)

	c.Limit = getConfigVar(
		configCL.Limit, configENV.Limit, configFile.Limit, defaultLimit, 0)

	c.PublicCryptoKey = getConfigVar(
		configCL.PublicCryptoKey, configENV.PublicCryptoKey, configFile.PublicCryptoKey, defaultCryptoKeyPath, "")

	c.Path = path
}

// UnmarshalJSON - For anmarshaling of the time parameters of the configuration file.
func (c *ConfigAgent) UnmarshalJSON(data []byte) error {
	type ConfigAgentJSON struct {
		Server         string `json:"address,omitempty"`
		PollInterval   string `json:"poll_interval,omitempty"`
		ReportInterval string `json:"report_interval,omitempty"`
	}

	var v ConfigAgentJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("agent config unmarshal error, err: %w", err)
	}

	c.Server = v.Server
	pi, err := time.ParseDuration(v.PollInterval)
	if err != nil {
		return fmt.Errorf("cannot parse poll interval duration err: %w", err)
	}
	c.PollInterval = int(pi.Seconds())

	ri, err := time.ParseDuration(v.ReportInterval)
	if err != nil {
		return fmt.Errorf("cannot parse report interval duration err: %w", err)
	}
	c.ReportInterval = int(ri.Seconds())

	return nil
}

func (c *ConfigAgent) String() string {
	return fmt.Sprintf("Addres: %s, ReportInterval: %d, PollInterval: %d, Limit: %d, Path: %s",
		c.Server, c.ReportInterval, c.PollInterval, c.Limit, c.Path)
}

// readConfigAgentFromENV - reading env vars and returned config.
func readConfigAgentFromENV() (*ConfigAgent, error) {
	c := newConfigAgent()

	if err := env.Parse(c); err != nil {
		return nil, fmt.Errorf("env parse agent config err: %w", err)
	}

	var hashkey string
	envkey := os.Getenv(envHashKey)
	if envkey != "" {
		hashkey = envkey
	}

	c.Key = []byte(hashkey)
	return c, nil
}

// readConfigAgentFromCL - reading command line vars and returned config.
func readConfigAgentFromCL() *ConfigAgent {
	c := newConfigAgent()

	var hashkey string
	flag.StringVar(&c.Server, metcollAddressFlagName, defaultMetcollAddress, "address metcoll server")
	flag.StringVar(&c.Path, configFlagName, defaultConfigPath, "path to json config file")
	flag.IntVar(&c.ReportInterval, reportIntervalFlagName, defaultReportInterval, "report push interval")
	flag.IntVar(&c.PollInterval, pollIntervalFlagName, defaultPollInterval, "poll interval")
	flag.StringVar(&hashkey, hashKeyFlagName, defaultHashKey, "hash key for setting up request hash")
	flag.IntVar(&c.Limit, limitFlagName, defaultLimit, "worker limit")
	flag.StringVar(&c.PublicCryptoKey, cryptoKeyFlagName, defaultCryptoKeyPath, "path to publickey.pem")

	flag.Parse()

	return c
}

// readConfigAgentFromFile - reading config file and returned config.
func readConfigAgentFromFile(path string) (*ConfigAgent, error) {
	if path == "" {
		return newConfigAgent(), nil
	}

	byteValue, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when reading agent configuration file err: %w", err)
	}

	c := newConfigAgent()
	if err := json.Unmarshal(byteValue, &c); err != nil {
		return nil, fmt.Errorf("an error occurred when unmarshal the agent configuration file err: %w", err)
	}

	return c, nil
}

func readFile(path string) ([]byte, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when opening the agent configuration file err: %w", err)
	}
	defer func() {
		if err := configFile.Close(); err != nil {
			zap.S().Errorf("closing config file was failed, err: %w", err)
		}
	}()

	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when parse the agent configuration file err: %w", err)
	}

	return byteValue, nil
}

// getConfigVar - compares variables received from
// the application command line, environment variable, and configuration file.
//
// Environment variables have higher priority over command line variables and config file.
// Command line variables have higher priority over variables from config file.
func getConfigVar[val comparable](varCL, varENV, varFile, def, empty val) val {
	v := def

	if varFile != empty && varFile != def {
		v = varFile
	}

	if varCL != empty && varCL != def {
		v = varCL
	}

	if varENV != empty && varENV != def {
		v = varENV
	}

	return v
}
