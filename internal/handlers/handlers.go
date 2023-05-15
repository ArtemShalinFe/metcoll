package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

var errUnknowMetricType = errors.New("unknow metric type")
var errMetricNotFound = errors.New("metric not found")
var errUpdateMetricError = errors.New("cannot update metric")

type Handler struct {
	values *storage.MemStorage
}

func NewHandler() *Handler {
	return &Handler{
		values: storage.NewMemStorage(),
	}
}

func ChiRouter() *chi.Mux {

	h := NewHandler()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", http.HandlerFunc(h.UpdateMetric))
	r.Get("/value/{metricType}/{metricName}", http.HandlerFunc(h.GetMetric))
	r.Get("/", http.HandlerFunc(h.GetMetricList))

	return r
}

func (h *Handler) UpdateMetric(w http.ResponseWriter, r *http.Request) {

	k := chi.URLParam(r, "metricName")
	v := chi.URLParam(r, "metricValue")
	t := chi.URLParam(r, "metricType")

	if strings.TrimSpace(k) == "" {
		http.Error(w, "name metric is empty", http.StatusBadRequest)
		return
	}

	newValue, err := updateValue(h, k, v, t)
	if err != nil {
		if errors.Is(err, errUpdateMetricError) {
			http.Error(w, "Bad request", http.StatusBadRequest)
		} else if errors.Is(err, errUnknowMetricType) {
			http.Error(w, errUnknowMetricType.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s %v", k, newValue)))

}

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {

	k := chi.URLParam(r, "metricName")
	t := chi.URLParam(r, "metricType")

	value, err := getValue(h, k, t)
	if err != nil {

		if errors.Is(err, errMetricNotFound) {
			http.Error(w, errMetricNotFound.Error(), http.StatusNotFound)
		} else if errors.Is(err, errUnknowMetricType) {
			http.Error(w, errUnknowMetricType.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	_, err = w.Write([]byte(value))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("GetMetric error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

func (h *Handler) GetMetricList(w http.ResponseWriter, r *http.Request) {

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
	for _, v := range h.values.GetDataList() {
		list += fmt.Sprintf(`<p>%s</p>`, v)
	}

	io.WriteString(w, fmt.Sprintf(body, list))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

func getValue(h *Handler, metricName string, metricType string) (string, error) {

	m, err := metrics.GetMetric(metricType)
	if err != nil {
		log.Printf("getMetric error: %v", err)
		return "", errUnknowMetricType
	}

	value, have := m.Get(h.values, metricName)
	if !have {
		return "", errMetricNotFound
	}

	return value, nil
}

func updateValue(h *Handler, metricName string, metricValue string, metricType string) (string, error) {

	m, err := metrics.GetMetric(metricType)
	if err != nil {
		log.Printf("getMetric error: %v", err)
		return "", errUnknowMetricType
	}

	newValue, err := m.Update(h.values, metricName, metricValue)
	if err != nil {
		return "", errUpdateMetricError
	}

	return newValue, nil
}
