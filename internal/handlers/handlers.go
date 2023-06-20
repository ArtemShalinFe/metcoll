package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

type Handler struct {
	storage Storage
	logger  *zap.SugaredLogger
}

type Storage interface {
	GetInt64Value(ctx context.Context, key string) (int64, error)
	GetFloat64Value(ctx context.Context, key string) (float64, error)
	AddInt64Value(ctx context.Context, key string, value int64) (int64, error)
	SetFloat64Value(ctx context.Context, key string, value float64) (float64, error)
	GetDataList(ctx context.Context) ([]string, error)
	BatchSetFloat64Value(ctx context.Context, gauges map[string]float64) (map[string]float64, []error, error)
	BatchAddInt64Value(ctx context.Context, counters map[string]int64) (map[string]int64, []error, error)
	Ping(ctx context.Context) error
}

func NewHandler(s Storage, l *zap.SugaredLogger) *Handler {

	return &Handler{
		storage: s,
		logger:  l,
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

	ms, err := h.storage.GetDataList(ctx)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("GetMetricList error: %w", err)
	}

	list := ""
	for _, v := range ms {
		list += fmt.Sprintf(`<p>%s</p>`, v)
	}

	resp := []byte(fmt.Sprintf(body, list))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	addHashHeader(w, []byte(resp))
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(resp); err != nil {
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

	h.logger.Infof("Trying update %s metric %s with value: %s", m.MType, m.ID, m.String())

	if err := m.Update(ctx, h.storage); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

	resp := fmt.Sprintf("%s %s", m.ID, m.String())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	addHashHeader(w, []byte(resp))
	w.WriteHeader(http.StatusOK)

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

	h.logger.Infof("Trying update %s metric %s with value: %s", m.MType, m.ID, m.String())
	h.logger.Debugf("UpdateMetric body: %s", string(b))

	if err := m.Update(ctx, h.storage); err != nil {
		if !errors.Is(err, storage.ErrNoRows) {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			h.logger.Errorf("UpdateMetric error: %w", err)
			return
		}
	}

	h.logger.Infof("Metric was updated - %s new value: %s", m.ID, m.String())

	b, err = json.Marshal(&m)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric marshal to json error: %w", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	addHashHeader(w, b)

	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("UpdateMetric error: %w", err)
		return
	}

}

func (h *Handler) BatchUpdate(ctx context.Context, w http.ResponseWriter, body io.ReadCloser) {

	w.Header().Set("Content-Type", "application/json")

	var ms []*metrics.Metrics

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

	h.logger.Debugf("BatchUpdate body: %s", string(b))

	for _, m := range ms {

		if m.MType != metrics.CounterMetric && m.MType != metrics.GaugeMetric {
			http.Error(w, "Bad request", http.StatusBadRequest)
			h.logger.Infof("metric %s has unknow type: %s", m.ID, m.MType)
			return
		}

		if m.Delta == nil && m.Value == nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			h.logger.Infof("metric %s has nil delta and value", m.ID)
			return
		}

	}

	// Автотесты хотят, чтобы мы возвращали ошибку изменения каждой метрики
	// для этого протянул errs
	ums, errs, err := metrics.BatchUpdate(ctx, ms, h.storage)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		h.logger.Errorf("BatchUpdate update error: %w", err)
		return
	}

	for _, um := range ums {
		h.logger.Infof("Metric %s was updated. New value: %s", um.ID, um.String())
	}

	b, err = json.Marshal(&errs)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("BatchUpdate marshal to json error: %w", err)
		return
	}

	addHashHeader(w, b)
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write(b); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("BatchUpdate error: %w", err)
		return
	}

}

func (h *Handler) ReadMetricFromURL(ctx context.Context, w http.ResponseWriter, id string, mType string) {

	m, err := metrics.GetMetric(id, mType)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := m.Get(ctx, h.storage); err != nil {
		if errors.Is(err, storage.ErrNoRows) {
			http.Error(w, "metric not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			h.logger.Errorf("UpdateMetric error: %w", err)
			return
		}
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

	if err := m.Get(ctx, h.storage); err != nil {
		if errors.Is(err, storage.ErrNoRows) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			h.logger.Errorf("ReadMetric error: %w", err)
			return
		}
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

	if err := h.storage.Ping(ctx); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		h.logger.Errorf("Ping error: %w", err)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func addHashHeader(w http.ResponseWriter, b []byte) {

	hash := hmac.New(sha256.New, []byte("hashkey"))
	hash.Write(b)
	w.Header().Set("HashSHA256", fmt.Sprintf("%x", hash.Sum(nil)))

}
