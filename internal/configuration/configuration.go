package configuration

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	Database        string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

func Parse() (*Config, error) {

	var c Config

	flag.StringVar(&c.Address, "a", "localhost:8080", "server end point")
	flag.IntVar(&c.StoreInterval, "i", 300, "storage saving interval")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "path to metric file-storage")
	flag.BoolVar(&c.Restore, "r", true, "restore metrics from a file at server startup")
	flag.StringVar(&c.Database, "d", "", "database connection")
	flag.StringVar(&c.Key, "k", "", "hash key")

	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	return &c, nil

}
