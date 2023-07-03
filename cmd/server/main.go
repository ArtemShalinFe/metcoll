package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/handlers"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

const (
	timeoutServerShutdown = time.Second * 30
	timeoutShutdown       = time.Second * 60
)

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {

	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	zl, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot init zap-logger err: %w ", err)
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		return fmt.Errorf("cannot init middleware logger err: %w ", err)
	}

	cfg, err := configuration.Parse()
	if err != nil {
		return fmt.Errorf("parse config err: %w ", err)
	}
	l.Info("parsed server config: ", fmt.Sprintf("%+v", cfg))

	componentsErrs := make(chan error, 1)
	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	// GS logger
	wg.Add(1)
	go func(errs chan<- error) {
		defer wg.Done()
		<-ctx.Done()

		if err := l.Interrupt(); err != nil {
			componentsErrs <- fmt.Errorf("cannot flush buffered log entries err: %w", err)
		}
	}(componentsErrs)

	stg, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		return fmt.Errorf("storage init err: %w ", err)
	}

	// GS storage
	wg.Add(1)
	go func(errs chan<- error) {
		defer wg.Done()
		<-ctx.Done()

		if err := stg.Interrupt(); err != nil {
			componentsErrs <- fmt.Errorf("close storage failed err: %w", err)
		}
	}(componentsErrs)

	l.Info("attempt to launch at address: ", cfg.Address)
	s := metcoll.NewServer(cfg)
	s.Handler = handlers.NewRouter(ctx,
		handlers.NewHandler(stg, l.SugaredLogger),
		l.RequestLogger,
		s.RequestHashChecker,
		s.ResponceHashSetter,
		compress.CompressMiddleware)

	// GS server
	go func(errs chan<- error) {
		if err := s.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				errs <- fmt.Errorf("listen and serve err: %w", err)
			}
		}
	}(componentsErrs)

	l.Info("server running at address: ", cfg.Address)

	wg.Add(1)
	go func(errs chan<- error) {
		defer l.Info("server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()

		if err := s.Shutdown(shutdownTimeoutCtx); err != nil {
			errs <- fmt.Errorf("server shutdown err: %w", err)
		}
	}(componentsErrs)

	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		l.Error(err)
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		l.Fatal("gracefull shutdown was failed")
	}()

	return nil
}
