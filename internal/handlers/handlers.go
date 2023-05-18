package handlers

import (
	"errors"

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

func (h *Handler) UpdateMetric(metricName string, metricValue string, metricType string) (string, error) {

	m, err := metrics.GetMetric(metricType)
	if err != nil {
		return "", errUnknowMetricType
	}

	newValue, err := m.Update(h.values, metricName, metricValue)
	if err != nil {
		return "", errUpdateMetricError
	}

	return newValue, nil

}

func (h *Handler) GetMetric(metricName string, metricType string) (string, error) {

	m, err := metrics.GetMetric(metricType)
	if err != nil {
		return "", errUnknowMetricType
	}

	value, ok := m.Get(h.values, metricName)
	if !ok {
		return "", errMetricNotFound
	}

	return value, nil

}

func (h *Handler) GetMetricList() []string {

	return h.values.GetDataList()

}
