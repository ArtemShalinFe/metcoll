package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Handlers interface {
	UpdateMetric(metricName string, metricValue string, metricType string) (string, error)
	GetMetric(metricName string, metricType string) (string, error)
	GetMetricList() []string
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	RequestLogger(h http.Handler) http.Handler
}

func NewRouter(h Handlers, log Logger) *chi.Mux {

	router := chi.NewRouter()
	router.Use(log.RequestLogger)
	router.Use(middleware.Recoverer)

	router.Post("/update/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {

		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		metricType := chi.URLParam(r, "metricType")

		if strings.TrimSpace(metricName) == "" {
			http.Error(w, "name metric is empty", http.StatusBadRequest)
			return
		}

		newValue, err := h.UpdateMetric(metricName, metricValue, metricType)
		if err != nil {
			if errors.Is(err, errUpdateMetricError) {
				http.Error(w, "Bad request", http.StatusBadRequest)
			} else if errors.Is(err, errUnknowMetricType) {
				http.Error(w, errUnknowMetricType.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Error("UpdateMetric error: ", err.Error())
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err = w.Write([]byte(fmt.Sprintf("%s %v", metricName, newValue))); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Error("UpdateMetric error: ", err)
			return
		}

	})

	router.Get("/value/{metricType}/{metricName}", func(w http.ResponseWriter, r *http.Request) {

		metricName := chi.URLParam(r, "metricName")
		metricType := chi.URLParam(r, "metricType")

		value, err := h.GetMetric(metricName, metricType)

		if err != nil {

			if errors.Is(err, errMetricNotFound) {
				http.Error(w, errMetricNotFound.Error(), http.StatusNotFound)
			} else if errors.Is(err, errUnknowMetricType) {
				http.Error(w, errUnknowMetricType.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Error("GetMetric error: ", err)
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err = w.Write([]byte(value)); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Error("GetMetric error: ", err)
			return
		}

	})

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {

		body := `
		<html>
		<head>
			<title>Metric list</title>
		</head>
		<body>
			<h1>Metric list</h1>
			%s
		</body>
		</html>`

		list := ""
		for _, v := range h.GetMetricList() {
			list += fmt.Sprintf(`<p>%s</p>`, v)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(fmt.Sprintf(body, list))); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Error("GetMetricList error: ", err)
			return
		}

	})

	return router
}
