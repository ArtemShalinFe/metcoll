package metcoll

import (
	"context"
	"errors"
	"fmt"
	"net"
	reflect "reflect"
	"testing"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
	"github.com/ArtemShalinFe/metcoll/internal/metrics"
	gomock "go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/test/bufconn"
)

type dialer struct {
	lis *bufconn.Listener
}

func (d *dialer) bufDialer(context.Context, string) (net.Conn, error) {
	return d.lis.Dial()
}

func NewDialer(t *testing.T, stg Storage) (*dialer, error) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	cfg := &configuration.Config{}

	s, err := NewGRPCServer(stg, cfg, zap.S())
	if err != nil {
		t.Fatalf("an occured error when initial grpc server, err: %v", err)
	}
	srv := NewMetricService(stg, zap.S())

	RegisterMetcollServer(s, srv)
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("server exited with error: %v", err)
		}
	}()

	return &dialer{
		lis: lis,
	}, nil
}

func TestMetricService_ReadMetric(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockStorage(ctrl)
	stg.EXPECT().GetInt64Value(gomock.Any(), metricc).
		Return(int64(11), nil)

	stg.EXPECT().GetFloat64Value(gomock.Any(), metricg).
		Return(float64(31.1), nil)

	stg.EXPECT().GetInt64Value(gomock.Any(), metricc).
		Return(int64(0), errors.New("unknow int64 error"))

	stg.EXPECT().GetFloat64Value(gomock.Any(), metricg).
		Return(float64(0), errors.New("unknow float64 error"))

	d, err := NewDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	type args struct {
		metric *metrics.Metrics
	}
	tests := []struct {
		name         string
		args         args
		want         *ReadMetricResponse
		wantPBMetric *Metric
		wantErr      bool
	}{
		{
			name: "#1",
			args: args{
				metric: metrics.NewCounterMetric(metricc, 11),
			},
			want: &ReadMetricResponse{},
			wantPBMetric: &Metric{
				Id:    metricc,
				Type:  Metric_COUNTER,
				Delta: 11,
			},
			wantErr: false,
		},
		{
			name: "#2",
			args: args{
				metric: metrics.NewGaugeMetric(metricg, 31.1),
			},
			want: &ReadMetricResponse{},
			wantPBMetric: &Metric{
				Id:    metricg,
				Type:  Metric_GAUGE,
				Value: 31.1,
			},
			wantErr: false,
		},
		{
			name: "#3",
			args: args{
				metric: metrics.NewCounterMetric(metricc, 11),
			},
			want:    &ReadMetricResponse{},
			wantErr: true,
		},
		{
			name: "#4",
			args: args{
				metric: metrics.NewGaugeMetric(metricg, 31.1),
			},
			want:    &ReadMetricResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(d.bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()
			client := NewMetcollClient(conn)

			req := &ReadMetricRequest{Metric: convertPBMetric(tt.args.metric)}
			got, err := client.ReadMetric(ctx, req)
			if err != nil && !tt.wantErr {
				t.Errorf("response MetcollClient.ReadMetric() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr && tt.wantPBMetric != nil {
				if !reflect.DeepEqual(got.Metric, tt.wantPBMetric) {
					t.Errorf("response Metric not equals, got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMetricService_Update(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockStorage(ctrl)
	stg.EXPECT().AddInt64Value(gomock.Any(), metricc, int64(11)).
		Return(int64(11), nil)

	stg.EXPECT().AddInt64Value(gomock.Any(), metricc, int64(11)).
		Return(int64(22), nil)

	stg.EXPECT().SetFloat64Value(gomock.Any(), metricg, float64(31.1)).
		Return(float64(31.1), nil)

	stg.EXPECT().SetFloat64Value(gomock.Any(), metricg, float64(32.1)).
		Return(float64(0), errors.New("unknow error"))

	d, err := NewDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	type args struct {
		metric *metrics.Metrics
	}
	tests := []struct {
		name         string
		args         args
		want         *UpdateResponse
		wantPBMetric *Metric
		wantErr      bool
	}{
		{
			name: "#1",
			args: args{
				metric: metrics.NewCounterMetric(metricc, 11),
			},
			want: &UpdateResponse{},
			wantPBMetric: &Metric{
				Id:    metricc,
				Type:  Metric_COUNTER,
				Delta: 11,
			},
			wantErr: false,
		},
		{
			name: "#2",
			args: args{
				metric: metrics.NewCounterMetric(metricc, 11),
			},
			want: &UpdateResponse{},
			wantPBMetric: &Metric{
				Id:    metricc,
				Type:  Metric_COUNTER,
				Delta: 22,
			},
			wantErr: false,
		},
		{
			name: "#3",
			args: args{
				metric: metrics.NewGaugeMetric(metricg, 31.1),
			},
			want:    &UpdateResponse{},
			wantErr: false,
		},
		{
			name: "#4",
			args: args{
				metric: &metrics.Metrics{
					ID:    "someUnknowType",
					MType: "unknow",
				},
			},
			want:    &UpdateResponse{},
			wantErr: true,
		},
		{
			name: "#5",
			args: args{
				metric: metrics.NewGaugeMetric(metricg, 32.1),
			},
			want:    &UpdateResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(d.bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()
			client := NewMetcollClient(conn)

			req := &UpdateRequest{Metric: convertPBMetric(tt.args.metric)}
			got, err := client.Update(ctx, req)
			if err != nil && !tt.wantErr {
				t.Errorf("response MetcollClient.Update() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr && tt.wantPBMetric != nil {
				if !reflect.DeepEqual(got.Metric, tt.wantPBMetric) {
					t.Errorf("response Metric not equals, got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMetricService_Updates(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockStorage(ctrl)

	var ms []*metrics.Metrics
	ms = append(ms, metrics.NewCounterMetric(metricc, 1),
		metrics.NewGaugeMetric(metricg, 1.2))

	counters := make(map[string]int64)
	counters[metricc] = 1
	stg.EXPECT().BatchAddInt64Value(gomock.Any(), counters).Times(1).Return(counters, nil, nil)

	gauges := make(map[string]float64)
	gauges[metricg] = 1.2
	stg.EXPECT().BatchSetFloat64Value(gomock.Any(), gauges).Times(1).Return(gauges, nil, nil)

	var counterMetrics []*metrics.Metrics
	counterMetrics = append(counterMetrics, metrics.NewCounterMetric(metricc, 1))

	var gaugeMetrics []*metrics.Metrics
	gaugeMetrics = append(gaugeMetrics, metrics.NewGaugeMetric(metricg, 1.2))

	updatedCounterMetrics := make(map[string]int64)
	updatedCounterMetrics[metricc] = 1
	updatedGaugeMetrics := make(map[string]float64)
	updatedGaugeMetrics[metricg] = 1.2

	stg.EXPECT().BatchSetFloat64Value(gomock.Any(), updatedGaugeMetrics).Times(1).
		Return(nil, nil, errors.New("error batch update float64"))

	stg.EXPECT().BatchSetFloat64Value(gomock.Any(), make(map[string]float64)).Times(1).
		Return(make(map[string]float64), nil, nil)

	stg.EXPECT().BatchAddInt64Value(gomock.Any(), updatedCounterMetrics).Times(1).
		Return(nil, nil, errors.New("error batch update int64"))

	d, err := NewDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	type args struct {
		metrics []*metrics.Metrics
	}
	tests := []struct {
		name    string
		args    args
		want    *UpdateResponse
		wantErr bool
	}{
		{
			name: "#1",
			args: args{
				metrics: ms,
			},
			want:    &UpdateResponse{},
			wantErr: false,
		},
		{
			name: "#2",
			args: args{
				metrics: gaugeMetrics,
			},
			want:    &UpdateResponse{},
			wantErr: true,
		},
		{
			name: "#3",
			args: args{
				metrics: counterMetrics,
			},
			want:    &UpdateResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(d.bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()
			client := NewMetcollClient(conn)

			var req BatchUpdateRequest
			for _, mtrs := range tt.args.metrics {
				pbm := convertPBMetric(mtrs)
				req.Metrics = append(req.Metrics, pbm)
			}

			got, err := client.Updates(ctx, &req)
			if err != nil && !tt.wantErr {
				t.Errorf("response MetcollClient.Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricService_MetricList(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	stg := NewMockStorage(ctrl)

	var data []string
	data = append(data, fmt.Sprintf("%s %d", metricc, 1), fmt.Sprintf("%s %f", metricg, 1.2))

	stg.EXPECT().GetDataList(gomock.Any()).Times(1).Return(data, nil)
	stg.EXPECT().GetDataList(gomock.Any()).Times(1).Return(nil, errors.New("any data list error"))

	d, err := NewDialer(t, stg)
	if err != nil {
		t.Errorf("an occured error when creating a new dialer, err: %v", err)
	}

	tests := []struct {
		name    string
		req     *MetricListRequest
		want    *MetricListResponse
		wantErr bool
	}{
		{
			name: "#1",
			req:  &MetricListRequest{},
			want: &MetricListResponse{
				Htmlpage: "\n\t<html>\n\t<head>\n\t\t<title>Metric list</title>\n\t</head>\n\t<body>\n\t\t<h1>Metric list</h1>\n\t\t<p>metricc 1</p><p>metricg 1.200000</p>\n\t</body>\n\t</html>",
			},
			wantErr: false,
		},
		{
			name: "#2",
			req:  &MetricListRequest{},
			want: &MetricListResponse{
				Htmlpage: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(d.bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Errorf("failed to dial bufnet: %v", err)
			}
			defer conn.Close()
			client := NewMetcollClient(conn)
			got, err := client.MetricList(ctx, tt.req)
			if err != nil && !tt.wantErr {
				t.Errorf("response MetcollClient.MetricList() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr && !reflect.DeepEqual(got.Htmlpage, tt.want.Htmlpage) {
				t.Errorf("response MetricList not equals, got = %v, want %v", got, tt.want)
			}
		})
	}
}
