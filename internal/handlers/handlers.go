package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
)

type Handler struct {
	values Storage
	logger Logger
}

type Storage interface {
	GetInt64Value(key string) (int64, bool)
	GetFloat64Value(key string) (float64, bool)
	AddInt64Value(key string, value int64) int64
	SetFloat64Value(key string, value float64) float64
	GetDataList() []string
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

func NewHandler(s Storage, l Logger) *Handler {

	return &Handler{
		values: s,
		logger: l,
	}
}

func (h *Handler) update(m *metrics.Metrics) error {

	return m.Update(h.values)

}

func (h *Handler) get(m *metrics.Metrics) bool {

	return m.Get(h.values)

}

func (h *Handler) CollectMetricList(w http.ResponseWriter) {

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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(fmt.Sprintf(body, list))); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("GetMetricList error: ", err)
	}

}

func (h *Handler) UpdateMetricFromURL(w http.ResponseWriter, id string, mType string, value string) {

	m, err := metrics.NewMetric(id, mType, value)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Error("UpdateMetric error: ", err.Error())
		return
	}

	if err := h.update(m); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("UpdateMetric error: ", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	resp := fmt.Sprintf("%s %s", m.ID, m.String())
	if _, err = w.Write([]byte(resp)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("UpdateMetric error: ", err)
		return
	}

	h.logger.Info("Metric was updated - ", m.ID, " new value: ", m.String())

}

func (h *Handler) UpdateMetric(w http.ResponseWriter, body io.ReadCloser) {

	var m metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("UpdateMetric read body error: ", err.Error())
		return
	}

	if err := json.Unmarshal(b, &m); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Error("UpdateMetric unmarshal error: ", err.Error())
		return
	}

	if m.Delta == nil && m.Value == nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := h.update(&m); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("UpdateMetric error: ", err.Error())
		return
	}

	b, err = json.Marshal(&m)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("UpdateMetric marshal to json error: ", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("UpdateMetric error: ", err)
		return
	}

	h.logger.Info("Metric was updated - ", m.ID, " new value: ", m.String())

}

func (h *Handler) ReadMetricFromURL(w http.ResponseWriter, id string, mType string) {

	m, err := metrics.GetMetric(id, mType)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if ok := h.get(m); !ok {
		http.Error(w, "metric not found", http.StatusNotFound)
		h.logger.Info("ReadMetric not found metric: ", m.ID)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write([]byte(m.String())); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("GetMetric error: ", err)
		return
	}

}

func (h *Handler) ReadMetric(w http.ResponseWriter, body io.ReadCloser) {

	var m metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("ReadMetric error: ", err.Error())
		return
	}

	if err := json.Unmarshal(b, &m); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Error("ReadMetric marshal error: ", err.Error())
		return
	}

	if ok := h.get(&m); !ok {
		http.Error(w, "metric not found", http.StatusNotFound)
		h.logger.Info("ReadMetric not found metric: ", m.ID)
		return
	}

	b, err = json.Marshal(m)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("ReadMetric marshal to json error: ", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write([]byte(b)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Error("GetMetric error: ", err)
		return
	}

}
