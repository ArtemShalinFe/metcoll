// Package main is used to store indicators of their aggregation and return on request.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/build"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

const (
	// TimeoutServerShutdown is waiting time until the server gorutins are completed.
	timeoutServerShutdown = time.Second * 30

	// TimeoutShutdown is waiting time until the rest of the gorutins are completed.
	timeoutShutdown = time.Second * 60
)

func main() {
	if err := run(); err != nil {
		zap.S().Fatalf("an occured fatal err: %w", err)
	}
}

func run() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGQUIT)

	defer cancelCtx()

	componentsErrs := make(chan error, 1)
	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	// init logger
	zl, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot init zap-logger err: %w ", err)
	}
	sl := zl.Sugar()

	sl.Info(build.NewBuild())

	// graceful shutdown logger
	wg.Add(1)
	go func(errs chan<- error) {
		defer wg.Done()
		<-ctx.Done()
		if err := sl.Sync(); err != nil {
			if runtime.GOOS == "darwin" {
				errs <- nil
			} else {
				errs <- fmt.Errorf("cannot flush buffered log entries err: %w", err)
			}
		}
	}(componentsErrs)

	// parse config
	cfg, err := configuration.Parse()
	if err != nil {
		return fmt.Errorf("parse config err: %w ", err)
	}
	sl.Info("parsed server config: ", fmt.Sprintf("%+v", cfg))

	// init storage
	stg, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		return fmt.Errorf("storage init err: %w ", err)
	}

	// graceful shutdown storage
	wg.Add(1)
	go func(errs chan<- error) {
		defer wg.Done()
		<-ctx.Done()

		if err := stg.Interrupt(); err != nil {
			errs <- fmt.Errorf("close storage failed err: %w", err)
		}
	}(componentsErrs)

	// init server
	sl.Info("attempt to launch server at address: ", cfg.Address)
	s, err := metcoll.InitServer(ctx, stg, cfg, sl)
	if err != nil {
		return fmt.Errorf("cannot init metcollserver, err: %w", err)
	}

	go func(errs chan<- error) {
		if err := s.ListenAndServe(); err != nil {
			errs <- fmt.Errorf("listen and serve err: %w", err)
		}
	}(componentsErrs)
	sl.Info("server running at address: ", cfg.Address)

	// graceful shutdown server
	wg.Add(1)
	go func(errs chan<- error) {
		defer sl.Info("server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()

		if err := s.Shutdown(shutdownTimeoutCtx); err != nil {
			errs <- fmt.Errorf("server shutdown err: %w", err)
		}
	}(componentsErrs)

	// check errors
	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		sl.Error(err)
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		sl.Fatal("gracefull shutdown was failed")
	}()

	return nil
}
