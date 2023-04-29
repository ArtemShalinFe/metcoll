package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

	resp, err := http.Post(m.URIPathForPush(s), "text/plain", nil)
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer resp.Body.Close()
	userResult, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("Resp: [%d] [%s]", resp.StatusCode, string(userResult))
}
