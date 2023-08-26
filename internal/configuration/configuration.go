package configuration

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

const defaultStoreInterval = 300

// Config contains configuration for server.
type Config struct {
	Address         string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Database        string `env:"DATABASE_DSN"`
	Key             []byte
	StoreInterval   int  `env:"STORE_INTERVAL"`
	Restore         bool `env:"RESTORE"`
}

func Parse() (*Config, error) {
	var c Config

	var hashkey string

	flag.StringVar(&c.Address, "a", defaultMetcollAddress, "server end point")
	flag.IntVar(&c.StoreInterval, "i", defaultStoreInterval, "storage saving interval")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "path to metric file-storage")
	flag.BoolVar(&c.Restore, "r", true, "restore metrics from a file at server startup")
	flag.StringVar(&c.Database, "d", "", "database connection")
	flag.StringVar(&hashkey, "k", "", "hash key")

	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("env parse server config err: %w", err)
	}

	envkey := os.Getenv("KEY")
	if envkey != "" {
		hashkey = envkey
	}

	c.Key = []byte(hashkey)

	return &c, nil
}

func (c *Config) String() string {
	return fmt.Sprintf("Addres: %s, StoreInterval: %d, Restore: %t, DSN: %s, FS path: %s",
		c.Address, c.StoreInterval, c.Restore, c.Database, c.FileStoragePath)
}
