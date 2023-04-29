package main

import (
	"fmt"
	"net/http"
)

const gauge = "gauge"
const counter = "counter"

type Server struct {
	host string
	port string
}
type Metric struct {
	name  string
	value string
	mType string
}

func NewMetric(Name string, Value string, mType string) *Metric {
	return &Metric{
		name:  Name,
		value: Value,
		mType: mType,
	}
}

func (m *Metric) URIPathForPush(s *Server) string {
	return fmt.Sprintf("http://%s:%s/update/%s/%s/%s", s.host, s.port, m.mType, m.name, m.value)
}

func (m *Metric) Push(s *Server) {

	http.Post(m.URIPathForPush(s), "text/plain", nil)

}
