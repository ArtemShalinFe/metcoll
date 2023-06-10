package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewRouter(ctx context.Context, handlers *Handler, middlewares ...func(http.Handler) http.Handler) *chi.Mux {

	router := chi.NewRouter()
	router.Use(middlewares...)
	router.Use(middleware.Recoverer)

	router.Post("/update/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {

		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metricType := chi.URLParam(r, "metricType")

		if strings.TrimSpace(metricName) == "" {
			http.Error(w, "name metric is empty", http.StatusBadRequest)
			return
		}

		rctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handlers.UpdateMetricFromURL(rctx, w, metricName, metricType, metricValue)

	})

	router.Post("/update/", func(w http.ResponseWriter, r *http.Request) {

		rctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handlers.UpdateMetric(rctx, w, r.Body)

	})

	router.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {

		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		if strings.TrimSpace(metricName) == "" {
			http.Error(w, "name metric is empty", http.StatusBadRequest)
			return
		}

		rctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handlers.ReadMetricFromURL(rctx, w, metricName, metricType)

	})

	router.Post("/value/", func(w http.ResponseWriter, r *http.Request) {

		rctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handlers.ReadMetric(rctx, w, r.Body)

	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {

		rctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handlers.CollectMetricList(rctx, w)

	})

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {

		rctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handlers.Ping(rctx, w)

	})

	router.Post("/updates/", func(w http.ResponseWriter, r *http.Request) {

		rctx, cancel := context.WithCancel(ctx)
		defer cancel()

		handlers.BatchUpdate(rctx, w, r.Body)

	})

	return router

}
