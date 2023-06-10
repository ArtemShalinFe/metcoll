package handlers

import (
	"context"
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
	GetInt64Value(ctx context.Context, key string) (int64, bool)
	GetFloat64Value(ctx context.Context, key string) (float64, bool)
	AddInt64Value(ctx context.Context, key string, value int64) int64
	SetFloat64Value(ctx context.Context, key string, value float64) float64
	GetDataList(ctx context.Context) []string
	Ping(ctx context.Context) error
}

type Logger interface {
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

func NewHandler(s Storage, l Logger) *Handler {

	return &Handler{
		values: s,
		logger: l,
	}
}

func (h *Handler) CollectMetricList(ctx context.Context, w http.ResponseWriter) {

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

	for _, v := range h.values.GetDataList(ctx) {
		list += fmt.Sprintf(`<p>%s</p>`, v)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(fmt.Sprintf(body, list))); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("GetMetricList error: %w", err)
	}

}

func (h *Handler) UpdateMetricFromURL(ctx context.Context, w http.ResponseWriter, id string, mType string, value string) {

	m, err := metrics.NewMetric(id, mType, value)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

	if err := m.Update(ctx, h.values); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	resp := fmt.Sprintf("%s %s", m.ID, m.String())
	if _, err = w.Write([]byte(resp)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

	h.logger.Infof("Metric was updated - %s new value: %s", m.ID, m.String())

}

func (h *Handler) UpdateMetric(ctx context.Context, w http.ResponseWriter, body io.ReadCloser) {

	var m metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric read body error: %w", err)
		return
	}

	if err := json.Unmarshal(b, &m); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Errorf("UpdateMetric unmarshal error: %w", err)
		return
	}

	if m.MType != metrics.CounterMetric && m.MType != metrics.GaugeMetric {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if m.Delta == nil && m.Value == nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := m.Update(ctx, h.values); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

	b, err = json.Marshal(&m)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric marshal to json error: %w", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

	h.logger.Infof("Metric was updated - %s new value: %s", m.ID, m.String())

}

func (h *Handler) BatchUpdate(ctx context.Context, w http.ResponseWriter, body io.ReadCloser) {

	var ms []metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("BatchUpdate read body error: %w", err)
		return
	}

	if err := json.Unmarshal(b, &ms); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Errorf("BatchUpdate unmarshal error: %w", err)
		return
	}

	h.logger.Infof("BatchUpdate body: %s", string(b))

	var errs []error

	for _, m := range ms {

		if m.MType != metrics.CounterMetric && m.MType != metrics.GaugeMetric {
			errs = append(errs, fmt.Errorf("metric %s has unknow type: %s", m.ID, m.MType))
			continue
		}

		if m.Delta == nil && m.Value == nil {
			errs = append(errs, fmt.Errorf("metric %s has nil delta and value", m.ID))
			continue
		}

		if err := m.Update(ctx, h.values); err != nil {
			errs = append(errs, fmt.Errorf("cannot update metric %s", m.ID))
			h.logger.Errorf("BatchUpdate error: %w", err)
			continue
		}

		h.logger.Infof("Metrics was updated - %s new value: %s", m.ID, m.String())

	}

	b, err = json.Marshal(&errs)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("BatchUpdate marshal to json error: %w", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(errs) > 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if _, err = w.Write(b); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

}

func (h *Handler) ReadMetricFromURL(ctx context.Context, w http.ResponseWriter, id string, mType string) {

	m, err := metrics.GetMetric(id, mType)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if ok := m.Get(ctx, h.values); !ok {
		http.Error(w, "metric not found", http.StatusNotFound)
		h.logger.Infof("ReadMetric not found metric: %s", m.ID)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write([]byte(m.String())); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("GetMetric error: %w", err)
		return
	}

}

func (h *Handler) ReadMetric(ctx context.Context, w http.ResponseWriter, body io.ReadCloser) {

	var m metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("ReadMetric error: %w", err)
		return
	}

	if err := json.Unmarshal(b, &m); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Errorf("ReadMetric marshal error: %w", err)
		return
	}

	if m.MType != metrics.CounterMetric && m.MType != metrics.GaugeMetric {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if ok := m.Get(ctx, h.values); !ok {
		http.Error(w, "metric not found", http.StatusNotFound)
		h.logger.Infof("ReadMetric not found metric: %s", m.ID)
		return
	}

	b, err = json.Marshal(m)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("ReadMetric marshal to json error: %w", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write([]byte(b)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("GetMetric error: %w", err)
		return
	}

}

func (h *Handler) Ping(ctx context.Context, w http.ResponseWriter) {

	if err := h.values.Ping(ctx); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("Ping error: %w", err)
		return
	}

	w.WriteHeader(http.StatusOK)

}
