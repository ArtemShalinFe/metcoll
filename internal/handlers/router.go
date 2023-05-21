package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewRouter(hs *Handler, middlewares ...func(http.Handler) http.Handler) *chi.Mux {

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

		hs.UpdateMetricFromURL(w, metricName, metricType, metricValue)

	})

	router.Post("/update/", func(w http.ResponseWriter, r *http.Request) {

		hs.UpdateMetric(w, r.Body)

	})

	router.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {

		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		hs.ReadMetricFromURL(w, metricName, metricType)

	})

	router.Post("/value/", func(w http.ResponseWriter, r *http.Request) {

		hs.ReadMetric(w, r.Body)

	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {

		hs.CollectMetricList(w)

	})

	return router

}
