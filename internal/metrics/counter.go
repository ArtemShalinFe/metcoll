package metrics

import (
	"fmt"
	"strconv"
)

type Counter struct{}

func (c *Counter) Update(values Storage, k string, v string) (string, error) {

	var newValue int64

	parsedValue, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fmt.Sprintf("%v", newValue), err
	}

	newValue = values.AddInt64Value(k, parsedValue)
	return strconv.FormatInt(newValue, 10), nil

}

func (c *Counter) Get(values Storage, k string) (string, bool) {

	value, have := values.GetInt64Value(k)
	return strconv.FormatInt(value, 10), have

}
