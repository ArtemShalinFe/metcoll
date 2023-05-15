package metrics

import (
	"errors"
	"strings"
)

const gaugeMetric = "gauge"
const counterMetric = "counter"

type Storage interface {
	GetInt64Value(key string) (int64, bool)
	GetFloat64Value(key string) (float64, bool)
	AddInt64Value(key string, value int64) int64
	SetFloat64Value(key string, value float64) float64
	GetCounterList() []string
	GetGaugeList() []string
}

type Metric interface {
	Update(values Storage, k string, v string) (string, error)
	Get(values Storage, k string) (string, bool)
}

func GetMetric(mType string) (Metric, error) {

	var m Metric

	if strings.ToLower(mType) == gaugeMetric {
		m = &Gauge{}
	} else if strings.ToLower(mType) == counterMetric {
		m = &Counter{}
	} else {
		return m, errors.New("unknow type metric")
	}

	return m, nil

}
