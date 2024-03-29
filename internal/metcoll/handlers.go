package metcoll

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

const (
	contentType     = "Content-Type"
	textPlain       = "text/plain; charset=utf-8"
	applicationJSON = "application/json"
	realIP          = "X-Real-IP"
)

type Handler struct {
	storage Storage
	logger  *zap.SugaredLogger
}

type Storage interface {
	// GetInt64Value - returns the metric value or ErrNoRows if it does not exist.
	GetInt64Value(ctx context.Context, key string) (int64, error)

	// GetFloat64Value - returns the metric value or ErrNoRows if it does not exist.
	GetFloat64Value(ctx context.Context, key string) (float64, error)

	// AddInt64Value - Saves the metric value for the key and returns the new metric value.
	AddInt64Value(ctx context.Context, key string, value int64) (int64, error)

	// SetFloat64Value - Saves the metric value for the key and returns the new metric value.
	SetFloat64Value(ctx context.Context, key string, value float64) (float64, error)

	// GetDataList - Returns all saved metrics.
	// Metric output format: <MetricName> <Value>
	//
	// Example:
	//
	//	MetricOne 1
	//	MetricTwo 2
	//	...
	GetDataList(ctx context.Context) ([]string, error)

	// BatchSetFloat64Value - Batch saving of metric values.
	// Returns the set metric values and errors for those metrics whose values could not be set.
	BatchSetFloat64Value(ctx context.Context, gauges map[string]float64) (map[string]float64, []error, error)

	// BatchAddInt64Value - Batch saving of metric values.
	// Returns the set metric values and errors for those metrics whose values could not be set.
	BatchAddInt64Value(ctx context.Context, counters map[string]int64) (map[string]int64, []error, error)

	Ping(ctx context.Context) error
}

func NewHandler(s Storage, l *zap.SugaredLogger) *Handler {
	return &Handler{
		storage: s,
		logger:  l,
	}
}

func templateMetricList() string {
	return `
	<html>
	<head>
		<title>Metric list</title>
	</head>
	<body>
		<h1>Metric list</h1>
		%s
	</body>
	</html>`
}

var mt = `<p>%s</p>`

func (h *Handler) CollectMetricList(ctx context.Context, w http.ResponseWriter) {
	ms, err := h.storage.GetDataList(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("an error occurred while getting metric list err: %w", err)
	}

	list := ""
	for _, v := range ms {
		list += fmt.Sprintf(mt, v)
	}

	resp := []byte(fmt.Sprintf(templateMetricList(), list))

	w.Header().Set(contentType, "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	h.writeResponseBody(w, resp)
}

func (h *Handler) UpdateMetricFromURL(ctx context.Context,
	w http.ResponseWriter, id string, mType string, value string) {
	m, err := metrics.NewMetric(id, mType, value)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Errorf("cannot get metric err: %w", err)
		return
	}

	h.logger.Infof("Trying update %s metric from URL %s with value: %s", m.MType, m.ID, m.String())

	if err := m.Update(ctx, h.storage); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("update metric in storage was failed err: %w", err)
		return
	}

	resp := fmt.Sprintf("%s %s", m.ID, m.String())

	w.Header().Set(contentType, textPlain)
	w.WriteHeader(http.StatusOK)
	h.writeResponseBody(w, []byte(resp))
}

func (h *Handler) UpdateMetric(ctx context.Context, w http.ResponseWriter, body io.ReadCloser) {
	var m metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("read body was failed err: %w", err)
		return
	}

	if err := json.Unmarshal(b, &m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Errorf("UpdateMetric unmarshal error: %w", err)
		return
	}

	if m.MType != metrics.CounterMetric && m.MType != metrics.GaugeMetric {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if m.Delta == nil && m.Value == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Infof("Trying update %s metric %s with value: %s", m.MType, m.ID, m.String())
	h.logger.Debugf("UpdateMetric body: %s", string(b))

	if err := m.Update(ctx, h.storage); err != nil {
		if !errors.Is(err, storage.ErrNoRows) {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Errorf("an error occurred while updating the metric error: %w", err)
			return
		}
	}

	b, err = json.Marshal(&m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("an error occurred while marshal to json the metric error: %w", err)
		return
	}
	w.Header().Set(contentType, applicationJSON)

	w.WriteHeader(http.StatusOK)
	h.writeResponseBody(w, b)
}

func (h *Handler) BatchUpdate(ctx context.Context, w http.ResponseWriter, body io.ReadCloser) {
	w.Header().Set(contentType, applicationJSON)

	var ms []*metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("BatchUpdate read body error: %w", err)
		return
	}

	if err := json.Unmarshal(b, &ms); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Errorf("BatchUpdate unmarshal error: %w", err)
		return
	}

	h.logger.Debugf("BatchUpdate body: %s", string(b))

	for _, m := range ms {
		if m.MType != metrics.CounterMetric && m.MType != metrics.GaugeMetric {
			w.WriteHeader(http.StatusBadRequest)
			h.logger.Infof("metric (%s) has unknow type: %s", m.ID, m.MType)
			return
		}

		if m.Delta == nil && m.Value == nil {
			w.WriteHeader(http.StatusBadRequest)
			h.logger.Infof("metric %s has nil delta and value", m.ID)
			return
		}
	}

	// Автотесты хотят, чтобы мы возвращали ошибку изменения каждой метрики
	// для этого протянул errs
	ums, errs, err := metrics.BatchUpdate(ctx, ms, h.storage)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Errorf("BatchUpdate update error: %w", err)
		return
	}

	for _, um := range ums {
		h.logger.Debugf("Metric (%s) was updated. New value: %s", um.ID, um.String())
	}

	b, err = json.Marshal(&errs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("BatchUpdate marshal to json error: %w", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.writeResponseBody(w, b)
}

func (h *Handler) ReadMetricFromURL(ctx context.Context, w http.ResponseWriter, id string, mType string) {
	m, err := metrics.GetMetric(id, mType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := m.Get(ctx, h.storage); err != nil {
		if errors.Is(err, storage.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Errorf("get metric from storage was failed err: %w", err)
			return
		}
	}

	w.Header().Set(contentType, textPlain)
	w.WriteHeader(http.StatusOK)
	h.writeResponseBody(w, []byte(m.String()))
}

func (h *Handler) ReadMetric(ctx context.Context, w http.ResponseWriter, body io.ReadCloser) {
	var m metrics.Metrics

	b, err := io.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("an error occurred while reading body error: %w", err)
		return
	}

	if err := json.Unmarshal(b, &m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.logger.Errorf("an error occurred while unmarshal body error: %w", err)
		return
	}

	if m.MType != metrics.CounterMetric && m.MType != metrics.GaugeMetric {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := m.Get(ctx, h.storage); err != nil {
		if errors.Is(err, storage.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Errorf("an error occurred while getting value error: %w", err)
			return
		}
	}

	b, err = json.Marshal(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("an error occurred while marshal metric to json error: %w", err)
		return
	}
	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(http.StatusOK)
	h.writeResponseBody(w, b)
}

func (h *Handler) Ping(ctx context.Context, w http.ResponseWriter) {
	if err := h.storage.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("Ping error: %w", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) writeResponseBody(w http.ResponseWriter, b []byte) {
	if _, err := w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorf("an error occurred while writing the response error: %w", err)
	}
}
