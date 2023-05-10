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

	flagAddress := flag.String("a", "localhost:8080", "server end point")
	flag.Parse()

	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}

	if c.Address == "" {
		c.Address = *flagAddress
	}

	return &c, nil

}
