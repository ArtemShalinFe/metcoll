package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/handlers"
	"github.com/ArtemShalinFe/metcoll/internal/interrupter"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

func main() {

	i := interrupter.NewInterrupters()

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	i.Use(l.Interrupt)

	cfg, err := configuration.Parse()
	if err != nil {
		l.Error("cannot parse server config file err: ", err)
		return
	}

	l.Info("parsed server config: ", fmt.Sprintf("%+v", cfg))

	ctx := context.Background()

	stg, err := storage.InitStorage(ctx, cfg, l)
	if err != nil {
		l.Error("cannot init storage err: ", err)
		return
	}

	i.Use(stg.Interrupt)

	s := metcoll.NewServer(cfg)
	i.Use(s.Interrupt)

	s.Handler = handlers.NewRouter(ctx, handlers.NewHandler(stg, l), l.RequestLogger, compress.CompressMiddleware)

	i.Run(l)

	l.Info("Try running on address: ", cfg.Address)
	if err := s.ListenAndServe(); err != nil {
		l.Error("ListenAndServe() err: ", err)
	}

}
