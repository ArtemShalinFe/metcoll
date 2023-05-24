package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/handlers"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/statesaver"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func main() {

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()

	cfg, err := parseConfig()
	if err != nil {
		l.Error("cannot parse server config file err: ", err)
		return
	}

	l.Info("parsed server config: ", fmt.Sprintf("%+v", cfg))

	s := storage.NewMemStorage()

	st, err := statesaver.NewState(s, l, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
	if err != nil {
		l.Error("cannot init state saver err: ", err)
		return
	}

	h := handlers.NewHandler(s, l, st)
	r := handlers.NewRouter(h, l.RequestLogger, compress.CompressMiddleware)

	l.Info("Try running on address: ", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		l.Error("ListenAndServe() err: ", err)
	}

}

func parseConfig() (*Config, error) {

	var c Config

	flag.StringVar(&c.Address, "a", "localhost:8080", "server end point")
	flag.IntVar(&c.StoreInterval, "i", 300, "storage saving interval")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "path to metric file-storage")
	flag.BoolVar(&c.Restore, "r", true, "restore metrics from a file at server startup")

	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	return &c, nil

}
