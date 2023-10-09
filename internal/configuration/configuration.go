package configuration

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env"
)

const (
	defaultStoreInterval  = 300
	storeIntervalFlagName = "i"
	envStoreInterval      = "STORE_INTERVAL"

	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultRestore         = true

	defaultPrivateCryptoKeyPath = ""

	envHashKey = "KEY"

	trustedSubnetFlagName = "t"
	defaultTrustedSubnet  = ""

	useProtobuffFlagName = "pb"
	defaultUseProtobuff  = false
)

func newConfig() *Config {
	return &Config{
		Address:         defaultMetcollAddress,
		FileStoragePath: defaultFileStoragePath,
		StoreInterval:   defaultStoreInterval,
		Restore:         defaultRestore,
	}
}

// Config contains configuration for server.
type Config struct {
	Address          string `env:"ADDRESS" json:"address"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"  json:"store_file"`
	Database         string `env:"DATABASE_DSN" json:"database_dsn"`
	ConfigFile       string `env:"CONFIG"`
	PrivateCryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	Key              []byte
	StoreInterval    int    `env:"STORE_INTERVAL" json:"store_interval"`
	Restore          bool   `env:"RESTORE" json:"restore"`
	TrustedSubnet    string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	UseProtobuff     bool   `env:"USE_PROTOBUFF" json:"use_protobuff"`
	CertFilePath     string `env:"CERTIFICATE" json:"certificate"`
}

// Parse - return parsed config.
//
// Environment variables have higher priority over command line variables and config file.
// Command line variables have higher priority over variables from config file.
func Parse() (*Config, error) {
	configCL := readConfigFromCL()

	configENV, err := readConfigFromENV()
	if err != nil {
		return nil, fmt.Errorf("an error occurred when reading the srv configuration env var, err: %w", err)
	}

	path := getConfigVar(configCL.ConfigFile, configENV.ConfigFile, "", "", "")
	configFile, err := readConfigFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when reading the srv configuration file, err: %w", err)
	}

	var c Config
	c.setFromConfigs(configCL, configENV, configFile, path)

	return &c, nil
}

// UnmarshalJSON - For anmarshaling of the time parameters of the configuration file.
func (c *Config) UnmarshalJSON(data []byte) error {
	type ConfigJSON struct {
		Address         string `json:"address"`
		FileStoragePath string `json:"store_file"`
		Database        string `json:"database_dsn"`
		StoreInterval   string `json:"store_interval"`
		HashKey         string `json:"hashkey"`
		Restore         bool   `json:"restore"`
		TrustedSubnet   string `json:"trusted_subnet"`
		UseProtobuff    bool   `json:"use_protobuff"`
		CertFilePath    string `json:"certificate"`
	}

	var v ConfigJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("server config unmarshal error, err: %w", err)
	}

	c.Address = v.Address
	c.FileStoragePath = v.FileStoragePath
	c.Restore = v.Restore
	c.Database = v.Database
	c.UseProtobuff = v.UseProtobuff
	c.TrustedSubnet = v.TrustedSubnet
	c.Key = []byte(v.HashKey)
	c.CertFilePath = v.CertFilePath

	si, err := time.ParseDuration(v.StoreInterval)
	if err != nil {
		return fmt.Errorf("cannot parse store interval duration err: %w", err)
	}
	c.StoreInterval = int(si.Seconds())

	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("Addres: %s, StoreInterval: %d, Restore: %t, DSN: %s, FS path: %s, Config: %s, Protobuff: %t",
		c.Address, c.StoreInterval, c.Restore, c.Database, c.FileStoragePath, c.ConfigFile, c.UseProtobuff)
}

// setFromConfigs -  sets configuration values from instances obtained
// from command line variables, environment variables, configuration file variables.
func (c *Config) setFromConfigs(configCL, configENV, configFile *Config, path string) {
	c.Address = getConfigVar(
		configCL.Address, configENV.Address, configFile.Address, defaultMetcollAddress, "")

	c.FileStoragePath = getConfigVar(
		configCL.FileStoragePath, configENV.FileStoragePath, configFile.FileStoragePath, defaultFileStoragePath, "")

	c.Database = getConfigVar(
		configCL.Database, configENV.Database, configFile.Database, "", "")

	c.StoreInterval = getConfigVar(
		configCL.StoreInterval, configENV.StoreInterval, configFile.StoreInterval, defaultStoreInterval, 0)

	c.Restore = getConfigVar(
		configCL.Restore, configENV.Restore, configFile.Restore, defaultRestore, true)

	c.PrivateCryptoKey = getConfigVar(
		configCL.PrivateCryptoKey, configENV.PrivateCryptoKey, configFile.PrivateCryptoKey, defaultPrivateCryptoKeyPath, "")

	c.TrustedSubnet = getConfigVar(
		configCL.TrustedSubnet, configENV.TrustedSubnet, configFile.TrustedSubnet, defaultTrustedSubnet, "")

	c.UseProtobuff = getConfigVar(
		configCL.UseProtobuff, configENV.UseProtobuff, configFile.UseProtobuff, defaultUseProtobuff, false)

	c.Key = getConfigByteVar(configCL.Key, configENV.Key, configFile.Key)

	c.CertFilePath = getConfigVar(configCL.CertFilePath, configENV.CertFilePath, configFile.CertFilePath, "", "")

	c.ConfigFile = path
}

// readConfigFromENV - reading env vars and returned config.
func readConfigFromENV() (*Config, error) {
	c := newConfig()

	if err := env.Parse(c); err != nil {
		return nil, fmt.Errorf("env parse srv config err: %w", err)
	}

	var hashkey string
	envkey := os.Getenv(envHashKey)
	if envkey != "" {
		hashkey = envkey
	}

	c.Key = []byte(hashkey)
	return c, nil
}

// readConfigFromCL - reading command line vars and returned config.
func readConfigFromCL() *Config {
	c := newConfig()

	var hashkey string
	flag.StringVar(&c.Address, metcollAddressFlagName, defaultMetcollAddress, "server endpoint")
	flag.IntVar(&c.StoreInterval, storeIntervalFlagName, defaultStoreInterval, "storage saving interval")
	flag.StringVar(&c.FileStoragePath, "f", defaultFileStoragePath, "path to metric file-storage")
	flag.BoolVar(&c.Restore, "r", defaultRestore, "restore metrics from a file at server startup")
	flag.StringVar(&c.Database, "d", "", "database connection")
	flag.StringVar(&hashkey, hashKeyFlagName, defaultHashKey, "hash key for check agent request hash")
	flag.StringVar(&c.PrivateCryptoKey, cryptoKeyFlagName, defaultCryptoKeyPath, "path to privatekey.pem")
	flag.StringVar(&c.TrustedSubnet, trustedSubnetFlagName, defaultTrustedSubnet, "trusted subnet, example 192.168.31.1")
	flag.BoolVar(&c.UseProtobuff, useProtobuffFlagName, defaultUseProtobuff, "use protobuf instead of http protocol")
	flag.StringVar(&c.CertFilePath, certFileFlagName, defaultCertFilePath, "absolute path to certificate (x509)")

	flag.Parse()

	c.Key = []byte(hashkey)

	return c
}

// readConfigFromFile - reading config file and returned config.
func readConfigFromFile(path string) (*Config, error) {
	if path == "" {
		return newConfig(), nil
	}

	byteValue, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf("an error occurred when reading srv configuration file err: %w", err)
	}

	c := newConfig()
	if err := json.Unmarshal(byteValue, &c); err != nil {
		return nil, fmt.Errorf("an error occurred when unmarshal the srv configuration file err: %w", err)
	}

	return c, nil
}
