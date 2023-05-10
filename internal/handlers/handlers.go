package handlers

import (
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

var values *storage.MemStorage

func ChiRouter() *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", http.HandlerFunc(UpdateMetric))
	r.Get("/value/{metricType}/{metricName}", http.HandlerFunc(GetMetric))
	r.Get("/", http.HandlerFunc(GetMetricList))

	return r
}

func UpdateMetric(w http.ResponseWriter, r *http.Request) {

	k := chi.URLParam(r, "metricName")
	v := chi.URLParam(r, "metricValue")

	if strings.TrimSpace(k) == "" {
		http.Error(w, "name metric is empty", http.StatusBadRequest)
		return
	}

	m, err := metrics.GetMetric(chi.URLParam(r, "metricType"))
	if err != nil {
		http.Error(w, "unknow type metric", http.StatusBadRequest)
		log.Printf("UpdateMetric error: %v", err)
		return
	}

	newValue, err := m.Update(values, k, v)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		log.Printf("UpdateMetric error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%s %v", k, newValue)))

}

func GetMetric(w http.ResponseWriter, r *http.Request) {

	k := chi.URLParam(r, "metricName")

	m, err := metrics.GetMetric(chi.URLParam(r, "metricType"))
	if err != nil {
		http.Error(w, "unknow type metric", http.StatusBadRequest)
		log.Printf("UpdateMetric error: %v", err)
		return
	}

	value, have := m.Get(values, k)
	if !have {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("UpdateMetric error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

func GetMetricList(w http.ResponseWriter, r *http.Request) {

	ml := values.GetDataList()

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
	for k, v := range ml {
		list += fmt.Sprintf(`<p>%s %v</p>`, k, v)
	}
	io.WriteString(w, fmt.Sprintf(body, list))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

func init() {
	values = storage.NewMemStorage()
}
