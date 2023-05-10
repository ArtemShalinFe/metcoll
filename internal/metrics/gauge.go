package metrics

import (
	"fmt"
	"strconv"
)

type Gauge struct{}

func (g *Gauge) Update(values Storage, k string, v string) (string, error) {

	var newValue float64

	parsedValue, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Sprintf("%v", newValue), err
	}

	newValue = values.SetFloat64Value(k, parsedValue)
	return fmt.Sprintf("%v", newValue), nil

}

func (g *Gauge) Get(values Storage, k string) (string, bool) {

	value, have := values.GetFloat64Value(k)
	return fmt.Sprintf("%v", value), have

}
