package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	MemStorage "github.com/ArtemShalinFe/metcoll/internal/storage"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const gaugeMetric = "gauge"
const counterMetric = "counter"

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

	if isGaugeMetric(r) {
		gaugeHandler(w, r)
		return
	}

	if isCounterMetric(r) {
		counterHandler(w, r)
		return
	}

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func GetMetric(w http.ResponseWriter, r *http.Request) {

	metricName := metricName(r)

	if isGaugeMetric(r) {

		have := MemStorage.Values.StorageHaveMetric(metricName)
		if !have {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		value, err := MemStorage.Values.GetFloat64Value(metricName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, fmt.Sprintf("%v", value))
	} else if isCounterMetric(r) {

		have := MemStorage.Values.StorageHaveMetric(metricName)
		if !have {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		value, err := MemStorage.Values.GetInt64Value(metricName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, fmt.Sprintf("%v", value))
	} else {
		http.Error(w, "Not implemented", http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetMetricList(w http.ResponseWriter, r *http.Request) {

	ml := MemStorage.Values.GetDataList()

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

func counterHandler(w http.ResponseWriter, r *http.Request) {

	k := metricName(r)
	v := metricValue(r)

	if strings.TrimSpace(k) == "" {
		http.Error(w, "name metric is empty", http.StatusBadRequest)
		return
	}

	parseValue, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	storageValue, err := MemStorage.Values.GetInt64Value(k)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newValue := parseValue + storageValue
	MemStorage.Values.SetInt64Value(k, newValue)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, fmt.Sprintf("%s %v", k, newValue))
}

func gaugeHandler(w http.ResponseWriter, r *http.Request) {

	k := metricName(r)
	v := chi.URLParam(r, "metricValue")

	if strings.TrimSpace(k) == "" {
		http.Error(w, "name metric is empty", http.StatusBadRequest)
		return
	}

	newValue, err := strconv.ParseFloat(v, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	MemStorage.Values.SetFloat64Value(k, newValue)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, fmt.Sprintf("%s %f", k, newValue))

}

func isGaugeMetric(r *http.Request) bool {
	return strings.ToLower(metricType(r)) == gaugeMetric
}

func isCounterMetric(r *http.Request) bool {
	return strings.ToLower(metricType(r)) == counterMetric
}

func metricType(r *http.Request) string {
	mt := chi.URLParam(r, "metricType")
	return mt
}

func metricName(r *http.Request) string {
	return chi.URLParam(r, "metricName")
}

func metricValue(r *http.Request) string {
	return chi.URLParam(r, "metricValue")
}
