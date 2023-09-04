// Package main is used to collect metrics and send them to the server.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/build"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metcoll"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/stats"
)

// timeoutShutdown is waiting time until the rest of the gorutins are completed.
const timeoutShutdown = time.Second * 60

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

	sl.Info(build.Info())

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		return fmt.Errorf("cannot init middleware logger err: %w", err)
	}

	cfg, err := configuration.ParseAgent()
	if err != nil {
		return fmt.Errorf("cannot parse server config file err: %w", err)
	}
	l.Infof("parsed agent config: %+v", cfg)

	componentsErrs := make(chan error, 1)
	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	wg.Add(1)
	go func(errs chan<- error) {
		defer wg.Done()
		<-ctx.Done()

		if err := l.Interrupt(); err != nil {
			errs <- fmt.Errorf("cannot flush buffered log entries err: %w", err)
		}
	}(componentsErrs)

	rl, err := logger.NewRLLogger(sl)
	if err != nil {
		return fmt.Errorf("cannot init retry logger err: %w", err)
	}

	client := metcoll.NewClient(cfg, rl)
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
				l.Errorf("batch update metrics failed err: %w", err)
			} else {
				stats.ClearPollCount()
			}
		}
	}()

	l.Info("metcoll client starting")

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
