package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/handlers"
	"github.com/ArtemShalinFe/metcoll/internal/interrupter"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

func main() {

	zl, err := zap.NewProduction()
	if err != nil {
		log.Fatal(fmt.Errorf("cannot init zap-logger err: %w ", err))
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot init middleware logger err: %w ", err))
	}

	i := interrupter.NewInterrupters()
	i.Use(l.Interrupt)

	cfg, err := configuration.Parse()
	if err != nil {
		l.Error("cannot parse server config file err: ", err)
		return
	}

	l.Info("parsed server config: ", fmt.Sprintf("%+v", cfg))

	ctx := context.Background()

	stg, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		l.Error("cannot init storage err: ", err)
		return
	}

	i.Use(stg.Interrupt)

	s := metcoll.NewServer(cfg)
	i.Use(func() error {

		if err := s.Shutdown(ctx); err != nil {
			return err
		}

		return nil

	})

	s.Handler = handlers.NewRouter(ctx,
		handlers.NewHandler(stg, l.SugaredLogger),
		l.RequestLogger,
		compress.CompressMiddleware,
		s.RequestHashChecker)

	i.Run(l.SugaredLogger)

	l.Info("Try running on address: ", cfg.Address)
	if err := s.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			l.Error("ListenAndServe() err: ", err)
		}
	}

}
