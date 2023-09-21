package metrics

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// GaugeMetric - name of gauge metric.
const GaugeMetric = "gauge"

// CounterMetric - name of counter metric.
const CounterMetric = "counter"

// PollCount - count of successful report submissions.
const PollCount = "PollCount"

type Storage interface {
	GetInt64Value(ctx context.Context, key string) (int64, error)
	GetFloat64Value(ctx context.Context, key string) (float64, error)
	AddInt64Value(ctx context.Context, key string, value int64) (int64, error)
	SetFloat64Value(ctx context.Context, key string, value float64) (float64, error)
	BatchSetFloat64Value(ctx context.Context, gauges map[string]float64) (map[string]float64, []error, error)
	BatchAddInt64Value(ctx context.Context, counters map[string]int64) (map[string]int64, []error, error)
}

// Metrics - an indicator that reflects a particular characteristic.
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

// errUnknowMetricType - error occurs when a metric other than gauge or counter is passed.
var errUnknowMetricType = errors.New("unknow metric type")

// GetMetric - Constructor for creating Metric-objects.
func GetMetric(id string, mType string) (*Metrics, error) {
	var m Metrics
	m.ID = id

	switch strings.ToLower(mType) {
	case GaugeMetric:
		m.MType = GaugeMetric
	case CounterMetric:
		m.MType = CounterMetric
	default:
		return nil, errUnknowMetricType
	}

	return &m, nil
}

// NewMetric - Object constructor.
func NewMetric(id string, mType string, value string) (*Metrics, error) {
	switch strings.ToLower(mType) {
	case GaugeMetric:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("an occured error when parse float for metric: %w", err)
		}
		return NewGaugeMetric(id, parsedValue), nil
	case CounterMetric:
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("an occured error when parse int for metric, err: %w", err)
		}
		return NewCounterMetric(id, parsedValue), nil
	default:
		return nil, errUnknowMetricType
	}
}

// NewGaugeMetric - Object constructor.
func NewGaugeMetric(id string, value float64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: GaugeMetric,
		Value: &value,
	}
}

// NewCounterMetric - Object constructor.
func NewCounterMetric(id string, delta int64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: CounterMetric,
		Delta: &delta,
	}
}

// IsPollCount - checks ID and MType. Returned true if it is "PollCount" and "counter".
func (m *Metrics) IsPollCount() bool {
	return m.MType == CounterMetric && m.ID == PollCount
}

func (m *Metrics) String() string {
	switch m.MType {
	case GaugeMetric:
		return strconv.FormatFloat(*m.Value, 'G', 10, 64)
	case CounterMetric:
		return strconv.FormatInt(*m.Delta, 10)
	default:
		return errUnknowMetricType.Error()
	}
}

// Update - updates the metric value in the storage.
func (m *Metrics) Update(ctx context.Context, storage Storage) error {
	switch m.MType {
	case GaugeMetric:
		newValue, err := storage.SetFloat64Value(ctx, m.ID, *m.Value)
		if err != nil {
			return fmt.Errorf("cannot update gauge metric err: %w", err)
		}

		m.Value = &newValue
	case CounterMetric:
		newValue, err := storage.AddInt64Value(ctx, m.ID, *m.Delta)
		if err != nil {
			return fmt.Errorf("cannot update counter metric err: %w", err)
		}

		m.Delta = &newValue
	default:
		return errUnknowMetricType
	}

	return nil
}

// BatchUpdate - batch update of metric values in the repository.
func BatchUpdate(ctx context.Context, ms []*Metrics, storage Storage) ([]*Metrics, []error, error) {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	for _, m := range ms {
		switch m.MType {
		case GaugeMetric:
			gauges[m.ID] = *m.Value
		case CounterMetric:
			counters[m.ID] += *m.Delta
		default:
			continue
		}
	}

	var ums []*Metrics
	var errs []error
	updatedGauges, uerrs, err := storage.BatchSetFloat64Value(ctx, gauges)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot exec batch update gauge metrics err: %w", err)
	}
	for id, val := range updatedGauges {
		ums = append(ums, NewGaugeMetric(id, val))
	}
	errs = append(errs, uerrs...)

	updatedCounters, uerrs, err := storage.BatchAddInt64Value(ctx, counters)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot exec batch update counter metrics err: %w", err)
	}
	for id, val := range updatedCounters {
		ums = append(ums, NewCounterMetric(id, val))
	}
	errs = append(errs, uerrs...)

	return ums, errs, nil
}

// Get - retrieves the current metric value from storage.
func (m *Metrics) Get(ctx context.Context, storage Storage) error {
	switch m.MType {
	case GaugeMetric:
		newValue, err := storage.GetFloat64Value(ctx, m.ID)
		if err != nil {
			return fmt.Errorf("cannot get gauge metric %s err: %w", m.ID, err)
		}
		m.Value = &newValue

		return nil
	case CounterMetric:
		newValue, err := storage.GetInt64Value(ctx, m.ID)
		if err != nil {
			return fmt.Errorf("cannot get counter metric %s err: %w", m.ID, err)
		}
		m.Delta = &newValue

		return nil
	default:
		return nil
	}
}
