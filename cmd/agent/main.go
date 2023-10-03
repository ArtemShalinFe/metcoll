// Package main is used to collect metrics and send them to the server.
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
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

// timeoutShutdown is waiting time until the rest of the gorutins are completed.
const timeoutShutdown = time.Second * 60

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

	zl, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot init zap-logger err: %w ", err)
	}
	sl := zl.Sugar()

	sl.Info(build.NewBuild())

	cfg, err := configuration.ParseAgent()
	if err != nil {
		return fmt.Errorf("cannot parse server config file err: %w", err)
	}
	sl.Infof("parsed agent config: %+v", cfg)

	componentsErrs := make(chan error, 1)
	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

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

	client, err := metcoll.InitClient(ctx, cfg, sl)
	if err != nil {
		return fmt.Errorf("cannot init metcoll client err: %w", err)
	}
	stats := stats.NewStats()

	go func() {
		mcs := make(chan []*metrics.Metrics, cfg.Limit)
		defer close(mcs)

		errs := make(chan error, cfg.Limit)
		defer close(errs)

		stats.RunCollectBatchStats(ctx, cfg, mcs)
		for i := 0; i < cfg.Limit; i++ {
			go client.BatchUpdateMetric(ctx, mcs, errs)
		}

		for err := range errs {
			if err != nil {
				sl.Errorf("batch update metrics failed err: %w", err)
			} else {
				stats.ClearPollCount()
			}
		}
	}()

	sl.Info("metcoll client starting")

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
