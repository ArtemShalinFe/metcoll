package metrics

import (
	"strconv"
)

type Gauge struct{}

func (g *Gauge) Update(values Storage, k string, v string) (string, error) {

	var newValue float64

	parsedValue, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return strconv.FormatFloat(newValue, 'G', 10, 64), err
	}

	newValue = values.SetFloat64Value(k, parsedValue)
	return strconv.FormatFloat(newValue, 'G', 10, 64), nil

}

func (g *Gauge) Get(values Storage, k string) (string, bool) {

	value, have := values.GetFloat64Value(k)
	return strconv.FormatFloat(value, 'G', 10, 64), have

}
