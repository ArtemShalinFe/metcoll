// Package handlers server REST API handlers for updating and retrieving metric values.
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

	router.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Use(middleware.Recoverer)

		r.Post("/update/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			metricName := chi.URLParam(r, "metricName")
			metricValue := chi.URLParam(r, "metricValue")
			metricType := chi.URLParam(r, "metricType")

			if strings.TrimSpace(metricName) == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			handlers.UpdateMetricFromURL(r.Context(), w, metricName, metricType, metricValue)
		})

		r.Post("/update/", func(w http.ResponseWriter, r *http.Request) {
			handlers.UpdateMetric(r.Context(), w, r.Body)
		})

		r.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {
			metricName := chi.URLParam(r, "metricName")
			metricType := chi.URLParam(r, "metricType")

			if strings.TrimSpace(metricName) == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			handlers.ReadMetricFromURL(r.Context(), w, metricName, metricType)
		})

		r.Post("/value/", func(w http.ResponseWriter, r *http.Request) {
			handlers.ReadMetric(r.Context(), w, r.Body)
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.CollectMetricList(r.Context(), w)
		})

		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			handlers.Ping(r.Context(), w)
		})

		r.Post("/updates/", func(w http.ResponseWriter, r *http.Request) {
			handlers.BatchUpdate(r.Context(), w, r.Body)
		})
	})

	router.Group(func(r chi.Router) {
		r.Mount("/debug", middleware.Profiler())
	})

	return router
}
