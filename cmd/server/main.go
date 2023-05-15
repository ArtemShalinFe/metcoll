package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/caarlos0/env"

	"github.com/ArtemShalinFe/metcoll/internal/handlers"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

func main() {

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	r := handlers.ChiRouter()

	log.Printf("Try running on %v\n", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		log.Fatal(err)
	}

}

func parseConfig() (*Config, error) {

	var c Config

	flag.StringVar(&c.Address, "a", "localhost:8080", "server end point")
	flag.Parse()

	if err := env.Parse(&c); err != nil {
		return nil, err
	}

	return &c, nil

}
