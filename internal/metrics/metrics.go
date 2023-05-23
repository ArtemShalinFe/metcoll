package metrics

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const GaugeMetric = "gauge"
const CounterMetric = "counter"
const PollCount = "PollCount"

type Storage interface {
	GetInt64Value(key string) (int64, bool)
	GetFloat64Value(key string) (float64, bool)
	AddInt64Value(key string, value int64) int64
	SetFloat64Value(key string, value float64) float64
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

var errUnknowMetricType = errors.New("unknow metric type")

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

func NewMetric(id string, mType string, value string) (*Metrics, error) {

	switch strings.ToLower(mType) {
	case GaugeMetric:

		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return NewGaugeMetric(id, parsedValue), nil

	case CounterMetric:

		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		return NewCounterMetric(id, parsedValue), nil

	default:

		return nil, errUnknowMetricType

	}

}

func NewGaugeMetric(id string, value float64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: GaugeMetric,
		Value: &value,
	}
}

func NewCounterMetric(id string, delta int64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: CounterMetric,
		Delta: &delta,
	}
}

func (m *Metrics) UnmarshalJSON(b []byte) error {

	type MetricAlias Metrics

	am := struct {
		MetricAlias
	}{
		MetricAlias: MetricAlias(*m),
	}

	if err := json.Unmarshal(b, &am); err != nil {
		return err
	}

	if am.MType != CounterMetric && am.MType != GaugeMetric {
		return errUnknowMetricType
	}

	m.ID = am.ID
	m.MType = am.MType
	m.Value = am.Value
	m.Delta = am.Delta

	return nil

}

func (m *Metrics) MarshalJSON() ([]byte, error) {

	type MetricAlias Metrics

	am := struct {
		MetricAlias
	}{
		MetricAlias: MetricAlias(*m),
	}

	if m.MType != CounterMetric && m.MType != GaugeMetric {
		return nil, errUnknowMetricType
	}

	return json.Marshal(am)

}

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

func (m *Metrics) Update(values Storage) error {

	switch m.MType {
	case GaugeMetric:

		newValue := values.SetFloat64Value(m.ID, *m.Value)
		m.Value = &newValue

	case CounterMetric:

		newValue := values.AddInt64Value(m.ID, *m.Delta)
		m.Delta = &newValue

	default:

		return errUnknowMetricType

	}

	return nil

}

func (m *Metrics) Get(values Storage) bool {

	switch m.MType {
	case GaugeMetric:

		newValue, ok := values.GetFloat64Value(m.ID)
		m.Value = &newValue
		return ok

	case CounterMetric:

		newValue, ok := values.GetInt64Value(m.ID)
		m.Delta = &newValue
		return ok

	default:

		return false

	}

}
