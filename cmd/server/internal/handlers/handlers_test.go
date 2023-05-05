package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	MemStorage "github.com/ArtemShalinFe/metcoll/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetric(t *testing.T) {

	ts := httptest.NewServer(ChiRouter())
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		status int
		method string
	}{
		{"/update/gauge/ /1.0", "name metric is empty\n", http.StatusBadRequest, http.MethodPost},
		{"/update/counter/ /1.0", "name metric is empty\n", http.StatusBadRequest, http.MethodPost},
		{"/update/gauge/metricg/1.0", "metricg 1.000000", http.StatusOK, http.MethodPost},
		{"/update/counter/metricc/1", "metricc 1", http.StatusOK, http.MethodPost},
		{"/update/gauge/", "404 page not found\n", http.StatusNotFound, http.MethodPost},
		{"/update/counter/", "404 page not found\n", http.StatusNotFound, http.MethodPost},
		{"/update/gauge/metric/novalue", "strconv.ParseFloat: parsing \"novalue\": invalid syntax\n", http.StatusBadRequest, http.MethodPost},
		{"/update/counter/metric/novalue", "strconv.ParseInt: parsing \"novalue\": invalid syntax\n", http.StatusBadRequest, http.MethodPost},
		{"/update/summary/metric/1", "Not implemented\n", http.StatusNotImplemented, http.MethodPost},
		{"/update/gauge/metricg/1.0", "", http.StatusMethodNotAllowed, http.MethodGet},
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

func TestGetMetric(t *testing.T) {

	MemStorage.Values.SetFloat64Value("metricg", 1.2)
	MemStorage.Values.SetFloat64Value("metrico", 2)
	MemStorage.Values.SetInt64Value("metricc", 1)

	ts := httptest.NewServer(ChiRouter())
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		status int
		method string
	}{
		{"/value/gauge/metricg", "1.2", http.StatusOK, http.MethodGet},
		{"/value/gauge/metrico", "2", http.StatusOK, http.MethodGet},
		{"/value/counter/metricc", "1", http.StatusOK, http.MethodGet},
		{"/value/gauge/", "404 page not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/counter/", "404 page not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/summary/metric", "Not implemented\n", http.StatusNotImplemented, http.MethodGet},
		{"/value/gauge/metricq", "", http.StatusMethodNotAllowed, http.MethodPost},
		{"/value/gauge/metricq", "Metric not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/counter/metricq", "Metric not found\n", http.StatusNotFound, http.MethodGet},
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

	MemStorage.Values.SetFloat64Value("metricg", 1.000000)
	MemStorage.Values.SetInt64Value("metricc", 1)

	ts := httptest.NewServer(ChiRouter())
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		status int
		method string
	}{
		{"/", "", http.StatusOK, http.MethodGet},
	}
	for _, v := range tests {
		resp, get := testRequest(t, ts, v.method, v.url)
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("URL: %s", v.url))
		assert.Contains(t, get, "metricg", "request home page")
		assert.Contains(t, get, "metricc", "request home page")
		assert.NotContains(t, get, "metricq", "request home page")
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
