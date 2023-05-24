package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/caarlos0/env"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/filestorage"
	"github.com/ArtemShalinFe/metcoll/internal/handlers"
	"github.com/ArtemShalinFe/metcoll/internal/interrupter"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

func main() {

	i := interrupter.NewInterrupters()

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	i.Use(l.LoggerInterrupt)

	cfg, err := parseConfig()
	if err != nil {
		l.Error("cannot parse server config file err: ", err)
		return
	}

	l.Info("parsed server config: ", fmt.Sprintf("%+v", cfg))

	h, err := initHandler(cfg, i, l)
	if err != nil {
		l.Error("cannot init handlers err: ", err)
		return
	}

	runGracefullInterrupt(l, i)

	r := handlers.NewRouter(h, l.RequestLogger, compress.CompressMiddleware)
	l.Info("Try running on address: ", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		l.Error("ListenAndServe() err: ", err)
	}

}

func initHandler(cfg *Config, i *interrupter.Interrupters, l *logger.AppLogger) (*handlers.Handler, error) {

	s := storage.NewMemStorage()
	if strings.TrimSpace(cfg.FileStoragePath) != "" {

		fs, err := filestorage.NewFilestorage(s, l, cfg.FileStoragePath, cfg.StoreInterval, cfg.Restore)
		if err != nil {
			l.Error("cannot init filestorage err: ", err)
			return nil, err
		}

		i.Use(fs.FilestorageInterrupt)

		l.Info("saving the state to a filestorage has been enabled")
		return handlers.NewHandler(fs, l), nil

	}

	l.Info("saving the state to a filestorage has been disabled - empty filestorage path")
	return handlers.NewHandler(s, l), nil

}

func runGracefullInterrupt(l *logger.AppLogger, i *interrupter.Interrupters) {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go func() {

		<-sigc

		ers := i.Do()

		if len(ers) > 0 {
			for _, v := range ers {
				l.Error(v)
			}
			os.Exit(1)
		}

		os.Exit(0)

	}()

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
