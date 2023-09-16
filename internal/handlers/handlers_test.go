package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/logger"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	"github.com/ArtemShalinFe/metcoll/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const (
	testServerInitErrTemplate = "an error occured while creating HTTP server for tests err: %v"
	bodyCloseErrTemplate      = "an error occured while body closing err: %v"
	temaplateURLErr           = "URL: %s"
	requestErrTemplate        = "http request err: %v"
	metricg                   = "metricg"
	metricc                   = "metricc"
)

func TestHandler_UpdateMetricFromURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().SetFloat64Value(gomock.Any(), metricg, float64(1.2)).Times(1).
		Return(float64(0), errors.New("failed update float64"))
	db.EXPECT().AddInt64Value(gomock.Any(), metricc, int64(1)).Times(1).
		Return(int64(0), errors.New("failed update int64"))

	mts, err := testServerWithMockStorage(db)
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer mts.Close()

	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		server *httptest.Server
		url    string
		want   string
		method string
		status int
		body   io.Reader
	}{
		{ts, "/update/gauge/metricg/1.2", "metricg 1.2", http.MethodPost, http.StatusOK, nil},
		{ts, "/update/counter/metricc/1", "metricc 1", http.MethodPost, http.StatusOK, nil},
		{ts, "/update/counter/ /1", "", http.MethodPost, http.StatusBadRequest, nil},
		{ts, "/update/gauge/", "", http.MethodPost, http.StatusNotFound, nil},
		{ts, "/update/counter/", "", http.MethodPost, http.StatusNotFound, nil},
		{ts, "/update/gauge/metric/novalue", "", http.MethodPost, http.StatusBadRequest, nil},
		{ts, "/update/counter/metric/novalue", "", http.MethodPost, http.StatusBadRequest, nil},
		{ts, "/update/summary/metric/1", "", http.MethodPost, http.StatusBadRequest, nil},
		{ts, "/update/gauge/metricg/1.0", "", http.MethodGet, http.StatusMethodNotAllowed, nil},
		{mts, "/update/gauge/metricg/1.2", "", http.MethodPost, http.StatusInternalServerError, nil},
		{mts, "/update/counter/metricc/1", "", http.MethodPost, http.StatusInternalServerError, nil},
	}
	for _, v := range tests {
		resp, get := testRequest(t, v.server, v.method, v.url, v.body)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().Ping(gomock.Any()).Return(errors.New("failed ping"))

	mts, err := testServerWithMockStorage(db)
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer mts.Close()

	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		server *httptest.Server
		url    string
		want   string
		method string
		status int
	}{
		{ts, "/ping", "", http.MethodGet, http.StatusOK},
		{mts, "/ping", "", http.MethodGet, http.StatusInternalServerError},
	}
	for _, v := range tests {
		resp, get := testRequest(t, v.server, v.method, v.url, nil)
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

func TestHandler_ReadMetricFromURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().GetFloat64Value(gomock.Any(), metricg).Times(1).
		Return(float64(1.2), nil)
	db.EXPECT().GetInt64Value(gomock.Any(), metricc).Times(1).
		Return(int64(1), nil)
	db.EXPECT().GetFloat64Value(gomock.Any(), metricg).Times(1).
		Return(float64(0), errors.New("failed to geting float64"))
	db.EXPECT().GetInt64Value(gomock.Any(), metricc).Times(1).
		Return(int64(0), errors.New("failed to geting int64"))

	mts, err := testServerWithMockStorage(db)
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer mts.Close()

	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		server *httptest.Server
		url    string
		want   string
		method string
		status int
		body   io.Reader
	}{
		{mts, "/value/gauge/metricg", "1.2", http.MethodGet, http.StatusOK, nil},
		{mts, "/value/counter/metricc", "1", http.MethodGet, http.StatusOK, nil},
		{ts, "/value/counter/ ", "", http.MethodGet, http.StatusBadRequest, nil},
		{ts, "/value/gauge/", "", http.MethodGet, http.StatusNotFound, nil},
		{ts, "/value/counter/", "", http.MethodGet, http.StatusNotFound, nil},
		{ts, "/value/summary/metric", "", http.MethodGet, http.StatusBadRequest, nil},
		{ts, "/value/gauge/metricq", "", http.MethodPost, http.StatusMethodNotAllowed, nil},
		{ts, "/value/gauge/metricq", "", http.MethodGet, http.StatusNotFound, nil},
		{ts, "/value/counter/metricq", "", http.MethodGet, http.StatusNotFound, nil},
		{mts, "/value/gauge/metricg", "", http.MethodGet, http.StatusInternalServerError, nil},
		{mts, "/value/counter/metricc", "", http.MethodGet, http.StatusInternalServerError, nil},
	}
	for _, v := range tests {
		resp, get := testRequest(t, v.server, v.method, v.url, v.body)
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

func TestHandler_UpdateMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().SetFloat64Value(gomock.Any(), metricg, float64(1.2)).Times(1).
		Return(float64(0), errors.New("failed update float64"))
	db.EXPECT().AddInt64Value(gomock.Any(), metricc, int64(1)).Times(1).
		Return(int64(0), errors.New("failed update int64"))

	mts, err := testServerWithMockStorage(db)
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer mts.Close()

	ts, err := testServer()
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer ts.Close()

	var tests = []struct {
		server      *httptest.Server
		bodyMetrics *metrics.Metrics
		want        *metrics.Metrics
		name        string
		url         string
		method      string
		status      int
	}{
		{
			server:      ts,
			name:        "#1",
			url:         "/update/",
			want:        metrics.NewGaugeMetric(metricg, 1.2),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewGaugeMetric(metricg, 1.2),
		},
		{
			server:      ts,
			name:        "#2",
			url:         "/update/",
			want:        metrics.NewGaugeMetric(metricg, 1.3),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewGaugeMetric(metricg, 1.3),
		},
		{
			server:      ts,
			name:        "#3",
			url:         "/update/",
			want:        metrics.NewCounterMetric(metricc, 1),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewCounterMetric(metricc, 1),
		},
		{
			server:      ts,
			name:        "#4",
			url:         "/update/",
			want:        metrics.NewCounterMetric(metricc, 2),
			status:      http.StatusOK,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewCounterMetric(metricc, 1),
		},
		{
			server:      ts,
			name:        "#5",
			url:         "/counter/",
			want:        &metrics.Metrics{},
			status:      http.StatusNotFound,
			method:      http.MethodPost,
			bodyMetrics: &metrics.Metrics{},
		},
		{
			server:      ts,
			name:        "#6",
			url:         "/update/",
			want:        &metrics.Metrics{},
			status:      http.StatusBadRequest,
			method:      http.MethodPost,
			bodyMetrics: &metrics.Metrics{},
		},
		{
			server:      ts,
			name:        "#7",
			url:         "/update/",
			want:        &metrics.Metrics{},
			status:      http.StatusMethodNotAllowed,
			method:      http.MethodGet,
			bodyMetrics: metrics.NewCounterMetric(metricc, 1),
		},
		{
			server: ts,
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
			server: ts,
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
			server:      mts,
			name:        "#10",
			url:         "/update/",
			want:        nil,
			status:      http.StatusInternalServerError,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewGaugeMetric(metricg, 1.2),
		},
		{
			server:      mts,
			name:        "#11",
			url:         "/update/",
			want:        nil,
			status:      http.StatusInternalServerError,
			method:      http.MethodPost,
			bodyMetrics: metrics.NewCounterMetric(metricc, 1),
		},
		{
			server:      ts,
			name:        "#12",
			url:         "/update/",
			want:        &metrics.Metrics{},
			status:      http.StatusBadRequest,
			method:      http.MethodPost,
			bodyMetrics: &metrics.Metrics{ID: metricc, MType: metrics.CounterMetric},
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

		resp, b := testRequest(t, v.server, v.method, v.url, bytes.NewBuffer(b))
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

func TestHandler_ReadMetric(t *testing.T) {
	const metricGaugeNotFound = "metricGaugeNotFound"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().GetInt64Value(gomock.Any(), metricc).Times(1).
		Return(int64(2), nil)
	db.EXPECT().GetFloat64Value(gomock.Any(), metricg).Times(1).
		Return(float64(1.3), nil)
	db.EXPECT().GetFloat64Value(gomock.Any(), metricGaugeNotFound).Times(1).
		Return(float64(0), storage.ErrNoRows)
	db.EXPECT().GetFloat64Value(gomock.Any(), metricg).Times(1).
		Return(float64(0), errors.New("failed update float64"))
	db.EXPECT().GetInt64Value(gomock.Any(), metricc).Times(1).
		Return(int64(0), errors.New("failed update int64"))

	mts, err := testServerWithMockStorage(db)
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer mts.Close()

	var tests = []struct {
		server      *httptest.Server
		bodyMetrics *metrics.Metrics
		want        *metrics.Metrics
		name        string
		url         string
		method      string
		status      int
	}{
		{
			server: mts,
			name:   "#1",
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
			server: mts,
			name:   "#2",
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
			server: mts,
			name:   "#3",
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
			server: mts,
			name:   "#4",
			url:    "/value/",
			want:   &metrics.Metrics{},
			status: http.StatusNotFound,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    metricGaugeNotFound,
				MType: metrics.GaugeMetric,
			},
		},
		{
			server: mts,
			name:   "#5",
			url:    "/value/",
			want:   &metrics.Metrics{},
			status: http.StatusInternalServerError,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    metricg,
				MType: metrics.GaugeMetric,
			},
		},
		{
			server: mts,
			name:   "#6",
			url:    "/value/",
			want:   &metrics.Metrics{},
			status: http.StatusInternalServerError,
			method: http.MethodPost,
			bodyMetrics: &metrics.Metrics{
				ID:    metricc,
				MType: metrics.CounterMetric,
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

		resp, b := testRequest(t, v.server, v.method, v.url, bytes.NewBuffer(b))
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

	var bodyMetricsErr1 []*metrics.Metrics
	bodyMetricsErr1 = append(bodyMetricsErr1, &metrics.Metrics{ID: "someid", MType: "wrongType"})

	var bodyMetricsErr2 []*metrics.Metrics
	bodyMetricsErr2 = append(bodyMetricsErr2, &metrics.Metrics{ID: "someid", MType: metrics.CounterMetric})

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
		{
			name:        "BatchUpdate #2",
			url:         "/updates/",
			want:        want,
			status:      http.StatusBadRequest,
			method:      http.MethodPost,
			bodyMetrics: bodyMetricsErr1,
		},
		{
			name:        "BatchUpdate #3",
			url:         "/updates/",
			want:        want,
			status:      http.StatusBadRequest,
			method:      http.MethodPost,
			bodyMetrics: bodyMetricsErr2,
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	db.EXPECT().GetDataList(gomock.Any()).AnyTimes().Return(nil, errors.New("any error"))

	mts, err := testServerWithMockStorage(db)
	if err != nil {
		t.Errorf(testServerInitErrTemplate, err)
	}
	defer mts.Close()

	var tests = []struct {
		server *httptest.Server
		url    string
		want   string
		method string
		status int
	}{
		{ts, "/update/gauge/metricg/1.2", "metricg 1.2", http.MethodPost, http.StatusOK},
		{ts, "/update/counter/metricc/1", "metricc 1", http.MethodPost, http.StatusOK},
		{ts, "/", "", http.MethodGet, http.StatusOK},
		{mts, "/", "", http.MethodGet, http.StatusInternalServerError},
	}
	for _, v := range tests {
		resp, get := testRequest(t, v.server, v.method, v.url, nil)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf(bodyCloseErrTemplate, err)
			}
		}()

		if v.method == http.MethodPost {
			continue
		}

		assert.Equal(t, v.status, resp.StatusCode, fmt.Sprintf(temaplateURLErr, v.url))
		if resp.StatusCode != http.StatusInternalServerError {
			var hp = "request home page"
			assert.NotContains(t, string(get), "metricq", hp)
			assert.Contains(t, string(get), metricg, hp)
			assert.Contains(t, string(get), metricc, hp)
		}
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

	sl := zap.L().Sugar()

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

func testServerWithMockStorage(s Storage) (*httptest.Server, error) {
	ctx := context.Background()

	sl := zap.L().Sugar()

	l, err := logger.NewMiddlewareLogger(sl)
	if err != nil {
		return nil, fmt.Errorf("cannot init middleware logger err: %w", err)
	}

	h := NewHandler(s, sl)
	r := NewRouter(ctx, h, l.RequestLogger)

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

type errRecorder struct{}

func (er *errRecorder) Header() http.Header {
	return nil
}

func (er *errRecorder) Write([]byte) (int, error) {
	return 0, errors.New("err recorder errors")
}

func (er *errRecorder) WriteHeader(statusCode int) {
}

func TestHandler_writeResponseBody(t *testing.T) {

	ctx := context.Background()
	cfg := &configuration.Config{}

	sl := zap.L().Sugar()

	s, err := storage.InitStorage(ctx, cfg, sl)
	if err != nil {
		t.Errorf("cannot init storage err: %v", err)
		return
	}

	type fields struct {
		storage Storage
		logger  *zap.SugaredLogger
	}
	type args struct {
		w http.ResponseWriter
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "#1",
			fields: fields{
				storage: s,
				logger:  sl,
			},
			args: args{
				w: httptest.NewRecorder(),
				b: []byte("useless string"),
			},
			wantErr: false,
		},
		{
			name: "#2",
			fields: fields{
				storage: s,
				logger:  sl,
			},
			args: args{
				w: &errRecorder{},
				b: []byte("useless string"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				storage: tt.fields.storage,
				logger:  tt.fields.logger,
			}
			h.writeResponseBody(tt.args.w, tt.args.b)
		})
	}
}
