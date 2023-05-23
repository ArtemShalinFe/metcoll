package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/caarlos0/env"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/handlers"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
	"github.com/ArtemShalinFe/metcoll/internal/storageStateSaver"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func main() {

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	ss := storageStateSaver.NewState(cfg.FileStoragePath)

	s, err := storage.NewMemStorage(cfg.Restore, cfg.StoreInterval, ss)
	if err != nil {
		log.Fatalf("cannot init mem storage err: %v", err)
	}

	h := handlers.NewHandler(s, l)
	r := handlers.NewRouter(h, l.RequestLogger, compress.CompressMiddleware)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go func() {

		sig := <-sigc
		log.Printf("incomming signal %v", sig)

		if err := s.SaveState(); err != nil {
			log.Printf("cannot save state err: %v", err)
		} else {
			log.Println("state was saved")
		}

		os.Exit(0)

	}()

	log.Printf("Try running on %v\n", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe() err: %v", err)
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
