package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
)

const (
	testServerInitErrTemplate = "an error occured while creating HTTP server for tests err: %v"
	bodyCloseErrTemplate      = "an error occured while body closing err: %v"
	temaplateURLErr           = "URL: %s"
	requestErrTemplate        = "http request err: %v"
)

func TestHandler_UpdateMetricFromURL(t *testing.T) {
	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		method string
		status int
	}{
		{"/update/gauge/metricg/1.2", "metricg 1.2", http.MethodPost, http.StatusOK},
		{"/update/counter/metricc/1", "metricc 1", http.MethodPost, http.StatusOK},
		{"/update/counter/ /1", "", http.MethodPost, http.StatusBadRequest},
		{"/update/gauge/", "", http.MethodPost, http.StatusNotFound},
		{"/update/counter/", "", http.MethodPost, http.StatusNotFound},
		{"/update/gauge/metric/novalue", "", http.MethodPost, http.StatusBadRequest},
		{"/update/counter/metric/novalue", "", http.MethodPost, http.StatusBadRequest},
		{"/update/summary/metric/1", "", http.MethodPost, http.StatusBadRequest},
		{"/update/gauge/metricg/1.0", "", http.MethodGet, http.StatusMethodNotAllowed},
		{"/value/gauge/metricg", "1.2", http.MethodGet, http.StatusOK},
		{"/value/counter/metricc", "1", http.MethodGet, http.StatusOK},
		{"/value/counter/ ", "", http.MethodGet, http.StatusBadRequest},
		{"/value/gauge/", "", http.MethodGet, http.StatusNotFound},
		{"/value/counter/", "", http.MethodGet, http.StatusNotFound},
		{"/value/summary/metric", "", http.MethodGet, http.StatusBadRequest},
		{"/value/gauge/metricq", "", http.MethodPost, http.StatusMethodNotAllowed},
		{"/value/gauge/metricq", "", http.MethodGet, http.StatusNotFound},
		{"/value/counter/metricq", "", http.MethodGet, http.StatusNotFound},
	}
	for _, v := range tests {
		resp, get := testRequest(t, ts, v.method, v.url, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf(bodyCloseErrTemplate, err)
			}
		}()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf(temaplateURLErr, v.url))
		if v.want != "" {
			assert.Equal(t, v.want, string(get), fmt.Sprintf(temaplateURLErr, v.url))
		}
	}
}

func TestHandler_Ping(t *testing.T) {
	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		method string
		status int
	}{
		{"/ping", "", http.MethodGet, http.StatusOK},
	}
	for _, v := range tests {
		resp, get := testRequest(t, ts, v.method, v.url, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf(bodyCloseErrTemplate, err)
			}
		}()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf(temaplateURLErr, v.url))
		if v.want != "" {
			assert.Equal(t, v.want, string(get), fmt.Sprintf(temaplateURLErr, v.url))
		}
	}
}

func ExampleHandler_UpdateMetricFromURL() {
	ts, err := testServer()
	if err != nil {
		fmt.Printf(testServerInitErrTemplate, err)
		return
	}
	defer ts.Close()

	res, err := ts.Client().Post(ts.URL+"/update/gauge/metricg/1.2", "plain/text", nil)
	if err != nil {
		fmt.Printf(requestErrTemplate, err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf(bodyCloseErrTemplate, err)
		}
	}()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

const metricg = "metricg"
const metricc = "metricc"

func TestHandler_UpdateMetric(t *testing.T) {
	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		bodyMetrics *metrics.Metrics
		want        *metrics.Metrics
		name        string
		url         string
		method      string
		status      int
	}{
		{
			name:        "#1",
			url:         "/update/",
			want:        metrics.NewGaugeMetric(metricg, 1.2),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewGaugeMetric(metricg, 1.2),
		},
		{
			name:        "#2",
			url:         "/update/",
			want:        metrics.NewGaugeMetric(metricg, 1.3),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewGaugeMetric(metricg, 1.3),
		},
		{
			name:        "#3",
			url:         "/update/",
			want:        metrics.NewCounterMetric(metricc, 1),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewCounterMetric(metricc, 1),
		},
		{
			name:        "#4",
			url:         "/update/",
			want:        metrics.NewCounterMetric(metricc, 2),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewCounterMetric(metricc, 1),
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
			bodyMetrics: metrics.NewCounterMetric(metricc, 1),
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
			want:   metrics.NewCounterMetric(metricc, 2),
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
			want:   metrics.NewGaugeMetric(metricg, 1.3),
			status: http.StatusOK,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    metricg,
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
				ID:    metricg,
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
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf(bodyCloseErrTemplate, err)
			}
		}()
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf(temaplateURLErr, v.url))

		if resp.StatusCode < 300 {
			var met metrics.Metrics
			if err = json.Unmarshal(b, &met); err != nil {
				t.Error(err)
			}
			require.Equal(t, v.want, &met, fmt.Sprintf(temaplateURLErr, v.url))
		}
	}
}

func ExampleHandler_UpdateMetric() {
	ts, err := testServer()
	if err != nil {
		fmt.Printf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	m := metrics.NewGaugeMetric(metricg, 1.2)
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("gauge metric marshal err : %w", err)
	}

	res, err := ts.Client().Post(ts.URL+"/update/", applicationJSON, bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf(requestErrTemplate, err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf(bodyCloseErrTemplate, err)
		}
	}()

	fmt.Println(res.StatusCode)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("example UpdateMetric read body error: %v", err)
		return
	}

	fmt.Println(string(bytes))

	// Output:
	// 200
	// {"value":1.2,"id":"metricg","type":"gauge"}
}

func TestHandler_BatchUpdate(t *testing.T) {
	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
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
		method      string
		bodyMetrics []*metrics.Metrics
		want        []string
		status      int
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
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf(bodyCloseErrTemplate, err)
			}
		}()
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

func ExampleHandler_BatchUpdate() {
	ts, err := testServer()
	if err != nil {
		fmt.Printf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var bodyMetrics []*metrics.Metrics
	bodyMetrics = append(bodyMetrics,
		metrics.NewCounterMetric("one", 1),
		metrics.NewCounterMetric("two", 2),
		metrics.NewGaugeMetric("three dot one", 3.1),
		metrics.NewGaugeMetric("four dot two", 4.2),
	)

	b, err := json.Marshal(bodyMetrics)
	if err != nil {
		fmt.Println("batch update marshal err : %w", err)
	}

	res, err := ts.Client().Post(ts.URL+"/updates/", applicationJSON, bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf(requestErrTemplate, err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf(bodyCloseErrTemplate, err)
		}
	}()

	fmt.Println(res.StatusCode)

	// Output:
	// 200
}

func TestHandler_CollectMetricList(t *testing.T) {
	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		url    string
		want   string
		method string
		status int
	}{
		{"/update/gauge/metricg/1.2", "metricg 1.2", http.MethodPost, http.StatusOK},
		{"/update/counter/metricc/1", "metricc 1", http.MethodPost, http.StatusOK},
		{"/", "", http.MethodGet, http.StatusOK},
	}
	for _, v := range tests {
		resp, get := testRequest(t, ts, v.method, v.url, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf(bodyCloseErrTemplate, err)
			}
		}()

		if v.method == http.MethodPost {
			continue
		}

		var hp = "request home page"
		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf(temaplateURLErr, v.url))
		assert.NotContains(t, string(get), "metricq", hp)
		assert.Contains(t, string(get), metricg, hp)
		assert.Contains(t, string(get), metricc, hp)
	}
}

func ExampleHandler_CollectMetricList() {
	ts, err := testServer()
	if err != nil {
		fmt.Printf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	res, err := ts.Client().Post(ts.URL+"/update/gauge/metric/1.2", textPlain, nil)
	if err != nil {
		fmt.Printf(requestErrTemplate, err)
		return
	}

	res, err = ts.Client().Post(ts.URL+"/update/counter/metric/1", textPlain, nil)
	if err != nil {
		fmt.Printf(requestErrTemplate, err)
		return
	}

	res, err = ts.Client().Post(ts.URL+"/update/counter/metric/1", textPlain, nil)
	if err != nil {
		fmt.Printf(requestErrTemplate, err)
		return
	}

	res, err = ts.Client().Get(ts.URL + "/")
	if err != nil {
		fmt.Printf(requestErrTemplate, err)
		return
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf(bodyCloseErrTemplate, err)
		}
	}()

	fmt.Println(res.StatusCode)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("example metric list read body error: %v", err)
		return
	}

	fmt.Println(string(bytes))

	// Output:
	// 200
	//
	//	<html>
	//	<head>
	//		<title>Metric list</title>
	//	</head>
	//	<body>
	//		<h1>Metric list</h1>
	//		<p>metric 1.2</p><p>metric 2</p>
	//	</body>
	//	</html>
}

func BenchmarkUpdateMetricFromURL(b *testing.B) {
	ctx := context.Background()
	cfg := &configuration.Config{}

	zl, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("cannot init zap-logger err: %v", err)
		return
	}
	sl := zl.Sugar()

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		fmt.Printf("cannot init logger err: %v", err)
		return
	}

	h := NewHandler(s, sl)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		id := strconv.Itoa(i)
		value := id

		b.StartTimer()
		h.UpdateMetricFromURL(ctx, httptest.NewRecorder(), id, metrics.CounterMetric, value)
	}
}

func BenchmarkUpdateMetric(b *testing.B) {
	ctx := context.Background()
	cfg := &configuration.Config{}

	zl, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("cannot init zap-logger err: %v", err)
		return
	}
	sl := zl.Sugar()

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		fmt.Printf("cannot init logger err: %v", err)
		return
	}

	h := NewHandler(s, sl)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		id := strconv.Itoa(i)

		gm := metrics.NewCounterMetric(id, int64(i))
		bs, err := json.Marshal(gm)
		if err != nil {
			fmt.Printf("marashal err: %v", err)
			return
		}
		body := io.NopCloser(bytes.NewReader(bs))

		b.StartTimer()
		h.UpdateMetric(ctx, httptest.NewRecorder(), body)
	}
}

func BenchmarkBatchUpdate(b *testing.B) {
	ctx := context.Background()
	cfg := &configuration.Config{}

	zl, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("cannot init zap-logger err: %v", err)
		return
	}
	sl := zl.Sugar()

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		fmt.Printf("cannot init logger err: %v", err)
		return
	}

	h := NewHandler(s, sl)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		id := strconv.Itoa(i)
		cm := metrics.NewCounterMetric(id, int64(i))

		var ms []*metrics.Metrics
		ms = append(ms, cm)
		for j := 0; j < 10; j++ {
			id := i + j
			gm := metrics.NewGaugeMetric(strconv.Itoa(id), float64(i+j))
			ms = append(ms, gm)
		}

		bs, err := json.Marshal(ms)
		if err != nil {
			fmt.Printf("marashal err: %v", err)
			return
		}
		body := io.NopCloser(bytes.NewReader(bs))

		b.StartTimer()
		h.BatchUpdate(ctx, httptest.NewRecorder(), body)
	}
}

func testServer() (*httptest.Server, error) {
	ctx := context.Background()
	cfg := &configuration.Config{}

	sl := zap.L().Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		return nil, fmt.Errorf("cannot init middleware logger err: %w", err)
	}

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		return nil, fmt.Errorf("cannot init storage err: %w", err)
	}

	h := NewHandler(s, sl)
	r := NewRouter(ctx, h, l.RequestLogger)

	h.CollectMetricList(ctx, httptest.NewRecorder())

	return httptest.NewServer(r), nil
}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader) (*http.Response, []byte) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf(bodyCloseErrTemplate, err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}
