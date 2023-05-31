package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/ArtemShalinFe/metcoll/internal/compress"
	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/handlers"
	"github.com/ArtemShalinFe/metcoll/internal/interrupter"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
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

	stg, err := storage.InitStorage(cfg, storage.NewMemStorage(), l)
	if err != nil {
		l.Error("cannot init storage err: ", err)
		return
	}

	i.Use(stg.Interrupt)

	s := NewHttpServer(cfg)
	s.Handler = handlers.NewRouter(handlers.NewHandler(stg, l), l.RequestLogger, compress.CompressMiddleware)

	runGracefullInterrupt(s, l, i)

	l.Info("Try running on address: ", cfg.Address)
	if err := s.ListenAndServe(); err != nil {
		l.Error("ListenAndServe() err: ", err)
	}

}

func NewHttpServer(cfg *configuration.Config) *http.Server {
	s := http.Server{
		Addr: cfg.Address,
	}
	return &s
}

func runGracefullInterrupt(s *http.Server, l *logger.AppLogger, i *interrupter.Interrupters) {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	go func() {

		<-sigc

		// пакет context же пока не проходили - пока todo поиспользую
		if err := s.Shutdown(context.TODO()); err != nil {
			l.Error(err)
		}

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
