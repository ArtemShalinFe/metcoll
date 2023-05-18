package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testLogger struct{}

func (tl *testLogger) Info(args ...any) {
	log.Println(args...)
}

func (tl *testLogger) Error(args ...any) {
	log.Println(args...)
}

func NewTestlogger() *testLogger {
	return &testLogger{}
}

func (tl *testLogger) RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		h.ServeHTTP(w, r)

	})
}

func TestUpdateMetric(t *testing.T) {

	l := NewTestlogger()
	h := NewHandler()
	r := NewRouter(h, l)

	ts := httptest.NewServer(r)
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		status int
		method string
	}{
		{"/update/gauge/metricg/1.2", "metricg 1.2", http.StatusOK, http.MethodPost},
		{"/update/counter/metricc/1", "metricc 1", http.StatusOK, http.MethodPost},
		{"/update/counter/ /1", "name metric is empty\n", http.StatusBadRequest, http.MethodPost},
		{"/update/gauge/", "404 page not found\n", http.StatusNotFound, http.MethodPost},
		{"/update/counter/", "404 page not found\n", http.StatusNotFound, http.MethodPost},
		{"/update/gauge/metric/novalue", "Bad request\n", http.StatusBadRequest, http.MethodPost},
		{"/update/counter/metric/novalue", "Bad request\n", http.StatusBadRequest, http.MethodPost},
		{"/update/summary/metric/1", "unknow metric type\n", http.StatusBadRequest, http.MethodPost},
		{"/update/gauge/metricg/1.0", "", http.StatusMethodNotAllowed, http.MethodGet},
		{"/value/gauge/metricg", "1.2", http.StatusOK, http.MethodGet},
		{"/value/counter/metricc", "1", http.StatusOK, http.MethodGet},
		{"/value/gauge/", "404 page not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/counter/", "404 page not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/summary/metric", "unknow metric type\n", http.StatusBadRequest, http.MethodGet},
		{"/value/gauge/metricq", "", http.StatusMethodNotAllowed, http.MethodPost},
		{"/value/gauge/metricq", "metric not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/counter/metricq", "metric not found\n", http.StatusNotFound, http.MethodGet},
	}
	for _, v := range tests {
		resp, get := testRequest(t, ts, v.method, v.url)
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("URL: %s", v.url))
		if v.want != "" {
			assert.Equal(t, v.want, get, fmt.Sprintf("URL: %s", v.url))
		}
	}
}

func TestGetMetricList(t *testing.T) {

	l := NewTestlogger()
	h := NewHandler()
	r := NewRouter(h, l)

	ts := httptest.NewServer(r)
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		status int
		method string
	}{
		{"/update/gauge/metricg/1.2", "metricg 1.2", http.StatusOK, http.MethodPost},
		{"/update/counter/metricc/1", "metricc 1", http.StatusOK, http.MethodPost},
		{"/", "", http.StatusOK, http.MethodGet},
	}
	for _, v := range tests {
		resp, get := testRequest(t, ts, v.method, v.url)
		defer resp.Body.Close()

		if v.method == http.MethodPost {
			continue
		}

		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("URL: %s", v.url))
		assert.NotContains(t, get, "metricq", "request home page")
		assert.Contains(t, get, "metricg", "request home page")
		assert.Contains(t, get, "metricc", "request home page")
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {

	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
