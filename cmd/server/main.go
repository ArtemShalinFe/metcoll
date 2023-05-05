package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	h "github.com/ArtemShalinFe/metcoll/cmd/server/internal/handlers"
	"github.com/caarlos0/env"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func main() {

	cfg := parseConfig()

	r := h.ChiRouter()

	fmt.Printf("Try running on %v\n", cfg.Address)
	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		panic(err)
	}

}

func parseConfig() *Config {

	var c Config

	flagAddress := flag.String("a", "", "server end point")
	flag.Parse()
	if *flagAddress != "" {
		c.Address = *flagAddress
		return &c
	}

	err := env.Parse(&c)
	if err != nil {
		log.Fatal(err)
	}
	return &c

}
