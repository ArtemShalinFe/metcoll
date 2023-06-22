package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

func TestUpdateMetricFromUrl(t *testing.T) {

	ctx := context.Background()
	cfg := &configuration.Config{}

	zl, err := zap.NewProduction()
	if err != nil {
		t.Errorf("cannot init zap-logger err: %v", err)
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		t.Errorf("cannot init middleware logger err: %v", err)
	}

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		t.Errorf("cannot init storage err: %v", err)
	}

	h := NewHandler(s, sl)
	r := NewRouter(ctx, h, l.RequestLogger)

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
		{"/update/summary/metric/1", "Bad request\n", http.StatusBadRequest, http.MethodPost},
		{"/update/gauge/metricg/1.0", "", http.StatusMethodNotAllowed, http.MethodGet},
		{"/value/gauge/metricg", "1.2", http.StatusOK, http.MethodGet},
		{"/value/counter/metricc", "1", http.StatusOK, http.MethodGet},
		{"/value/counter/ ", "name metric is empty\n", http.StatusBadRequest, http.MethodGet},
		{"/value/gauge/", "404 page not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/counter/", "404 page not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/summary/metric", "Bad request\n", http.StatusBadRequest, http.MethodGet},
		{"/value/gauge/metricq", "", http.StatusMethodNotAllowed, http.MethodPost},
		{"/value/gauge/metricq", "metric not found\n", http.StatusNotFound, http.MethodGet},
		{"/value/counter/metricq", "metric not found\n", http.StatusNotFound, http.MethodGet},
	}
	for _, v := range tests {
		resp, get := testRequest(t, ts, v.method, v.url, nil)
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("URL: %s", v.url))
		if v.want != "" {
			assert.Equal(t, v.want, string(get), fmt.Sprintf("URL: %s", v.url))
		}
	}
}

func TestUpdateMetric(t *testing.T) {

	ctx := context.Background()
	cfg := &configuration.Config{}

	zl, err := zap.NewProduction()
	if err != nil {
		t.Errorf("cannot init zap-logger err: %v", err)
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		t.Errorf("cannot init middleware logger err: %v", err)
	}

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		t.Errorf("cannot init logger err: %v", err)
	}

	h := NewHandler(s, sl)
	r := NewRouter(ctx, h, l.RequestLogger)

	ts := httptest.NewServer(r)
	defer ts.Close()

	var tests = []struct {
		name        string
		url         string
		status      int
		method      string
		bodyMetrics *metrics.Metrics
		want        *metrics.Metrics
	}{
		{
			name:        "#1",
			url:         "/update/",
			want:        metrics.NewGaugeMetric("metricg", 1.2),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewGaugeMetric("metricg", 1.2),
		},
		{
			name:        "#2",
			url:         "/update/",
			want:        metrics.NewGaugeMetric("metricg", 1.3),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewGaugeMetric("metricg", 1.3),
		},
		{
			name:        "#3",
			url:         "/update/",
			want:        metrics.NewCounterMetric("metricc", 1),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewCounterMetric("metricc", 1),
		},
		{
			name:        "#4",
			url:         "/update/",
			want:        metrics.NewCounterMetric("metricc", 2),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewCounterMetric("metricc", 1),
		},
		{
			name:        "#5",
			url:         "/counter/",
			want:        &metrics.Metrics{},
			status:      http.StatusNotFound,
			method:      http.MethodPost,
			bodyMetrics: &metrics.Metrics{},
		},
		{
			name:        "#6",
			url:         "/update/",
			want:        &metrics.Metrics{},
			status:      http.StatusBadRequest,
			method:      http.MethodPost,
			bodyMetrics: &metrics.Metrics{},
		},
		{
			name:        "#7",
			url:         "/update/",
			want:        &metrics.Metrics{},
			status:      http.StatusMethodNotAllowed,
			method:      http.MethodGet,
			bodyMetrics: metrics.NewCounterMetric("metricc", 1),
		},
		{
			name:   "#8",
			url:    "/update/",
			want:   &metrics.Metrics{},
			status: http.StatusBadRequest,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    "Alloc",
				MType: "HYPO",
			},
		},
		{
			name:   "#9",
			url:    "/update/",
			want:   &metrics.Metrics{},
			status: http.StatusBadRequest,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    "Alloc",
				MType: "HYPO",
				Value: new(float64),
			},
		},
		{
			name:   "#10",
			url:    "/value/",
			want:   metrics.NewCounterMetric("metricc", 2),
			status: http.StatusOK,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    "metricc",
				MType: metrics.CounterMetric,
			},
		},
		{
			name:   "#11",
			url:    "/value/",
			want:   metrics.NewGaugeMetric("metricg", 1.3),
			status: http.StatusOK,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    "metricg",
				MType: metrics.GaugeMetric,
			},
		},
		{
			name:   "#12",
			url:    "/value/",
			want:   &metrics.Metrics{},
			status: http.StatusBadRequest,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    "metricg",
				MType: "HYPE",
			},
		},
		{
			name:   "#13",
			url:    "/value/",
			want:   &metrics.Metrics{},
			status: http.StatusNotFound,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    "metricGaugeNotFound",
				MType: metrics.GaugeMetric,
			},
		},
	}

	type MetricAlias metrics.Metrics

	for _, v := range tests {

		am := struct {
			MetricAlias
		}{
			MetricAlias: MetricAlias(*v.bodyMetrics),
		}

		b, err := json.Marshal(am)
		if err != nil {
			t.Error(err)
		}

		resp, b := testRequest(t, ts, v.method, v.url, bytes.NewBuffer(b))
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("%s URL: %s", v.name, v.url))

		if resp.StatusCode < 300 {
			var met metrics.Metrics
			if err = json.Unmarshal(b, &met); err != nil {
				t.Error(err)
			}
			require.Equal(t, v.want, &met, fmt.Sprintf("%s URL: %s", v.name, v.url))
		}
	}
}

func TestHandler_BatchUpdate(t *testing.T) {

	ctx := context.Background()
	cfg := &configuration.Config{}

	zl, err := zap.NewProduction()
	if err != nil {
		t.Errorf("cannot init zap-logger err: %v", err)
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		t.Errorf("cannot init middleware logger err: %v", err)
	}

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		t.Errorf("cannot init logger err: %v", err)
	}

	h := NewHandler(s, sl)
	r := NewRouter(ctx, h, l.RequestLogger)

	ts := httptest.NewServer(r)
	defer ts.Close()

	var bodyMetrics []*metrics.Metrics
	bodyMetrics = append(bodyMetrics, metrics.NewCounterMetric("one", 1))
	bodyMetrics = append(bodyMetrics, metrics.NewCounterMetric("two", 2))
	bodyMetrics = append(bodyMetrics, metrics.NewGaugeMetric("three dot one", 3.1))
	bodyMetrics = append(bodyMetrics, metrics.NewGaugeMetric("four dot two", 4.2))

	var want []string

	var tests = []struct {
		name        string
		url         string
		status      int
		method      string
		bodyMetrics []*metrics.Metrics
		want        []string
	}{
		{
			name:        "BatchUpdate #1",
			url:         "/updates/",
			want:        want,
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: bodyMetrics,
		},
	}

	for _, v := range tests {

		b, err := json.Marshal(v.bodyMetrics)
		if err != nil {
			t.Error(err)
		}

		resp, b := testRequest(t, ts, v.method, v.url, bytes.NewBuffer(b))
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("%s URL: %s", v.name, v.url))

		if resp.StatusCode < 300 {
			var errs []string
			if err = json.Unmarshal(b, &errs); err != nil {
				t.Error(err)
			}
			require.Equal(t, v.want, errs, fmt.Sprintf("%s URL: %s", v.name, v.url))
		}
	}
}

func TestCollectMetricList(t *testing.T) {

	ctx := context.Background()
	cfg := &configuration.Config{}

	zl, err := zap.NewProduction()
	if err != nil {
		t.Errorf("cannot init zap-logger err: %v", err)
	}
	sl := zl.Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		t.Errorf("cannot init middleware logger err: %v", err)
	}

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		t.Errorf("cannot init logger err: %v", err)
	}

	h := NewHandler(s, sl)
	r := NewRouter(ctx, h, l.RequestLogger)

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
		resp, get := testRequest(t, ts, v.method, v.url, nil)
		defer resp.Body.Close()

		if v.method == http.MethodPost {
			continue
		}

		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("URL: %s", v.url))
		assert.NotContains(t, string(get), "metricq", "request home page")
		assert.Contains(t, string(get), "metricg", "request home page")
		assert.Contains(t, string(get), "metricc", "request home page")
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader) (*http.Response, []byte) {

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}
